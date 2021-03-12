package usecase_test

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/google/uuid"
	"github.com/m-mizutani/catbox/pkg/controller"
	"github.com/m-mizutani/catbox/pkg/db"
	"github.com/m-mizutani/catbox/pkg/interfaces"
	"github.com/m-mizutani/catbox/pkg/model"
	"github.com/m-mizutani/catbox/pkg/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrivyScanImage(t *testing.T) {
	t.Run("normal case", func(t *testing.T) {
		ctrl, mock := newControllerForTrivyScanImageTest(t)

		mock.s3.saveObject("example-bucket", "testing/cache/trivy/db/trivy.db", "dummy")
		mock.s3.saveObject("example-bucket", "testing/cache/trivy/db/metadata.json", "dummy")

		req := &model.ScanRequestMessage{
			RequestID:   uuid.New().String(),
			RequestedBy: "ecr.PushImage",
			RequestedAt: time.Now().UTC(),
			Target: model.TaggedImage{
				Registry: "111111111111.dkr.ecr.ap-northeast-1.amazonaws.com",
				Repo:     "strix",
				Tag:      "latest",
			},
			OutS3Prefix: "path/to/prefix/",
		}
		require.NoError(t, usecase.TrivyScanImage(ctrl, req))

		t.Run("can get saved ScanReport by repo", func(t *testing.T) {
			report, err := mock.dbClient.GetLatestScanReportsByRepo("111111111111.dkr.ecr.ap-northeast-1.amazonaws.com", "strix", "latest")
			require.NoError(t, err)
			require.NotNil(t, report)
			assert.Equal(t, req.RequestedAt.Unix(), report.RequestedAt)
			assert.Greater(t, report.StatusSeq, int64(0))
		})

		t.Run("report is saved on S3", func(t *testing.T) {
			require.Equal(t, 1, len(mock.s3.putObjectInput))
			assert.Contains(t, mock.s3.regions, "ap-northeast-0")
			assert.Equal(t, "example-bucket", *mock.s3.putObjectInput[0].Bucket)
			assert.Equal(t, "testing/path/to/prefix/trivy.json.gz", *mock.s3.putObjectInput[0].Key)
		})

		t.Run("sent an inspect request message to inspect queue", func(t *testing.T) {
			require.Equal(t, 1, len(mock.sqs.input))
			assert.Equal(t, "https://sqs.us-east-2.amazonaws.com/123456789012/inspect-queue", *mock.sqs.input[0].QueueUrl)

			var req model.InspectRequestMessage
			require.NoError(t, json.Unmarshal([]byte(*mock.sqs.input[0].MessageBody), &req))
			assert.NotEmpty(t, req.ReportID)
		})

		t.Run("accessed to registry with auth token by ecr.GetAuthorizationToken", func(t *testing.T) {
			require.Equal(t, 2, len(mock.http.requests))

			req1 := mock.http.requests[0]
			assert.Equal(t, "GET", req1.Method)
			assert.Equal(t, "https://111111111111.dkr.ecr.ap-northeast-1.amazonaws.com/v2/strix/manifests/latest", req1.URL.String())
			assert.Equal(t, "Basic fake_auth_token", req1.Header.Get("Authorization"))
			assert.Equal(t, "*/*", req1.Header.Get("Accept"))

			req2 := mock.http.requests[1]
			assert.Equal(t, "GET", req2.Method)
			assert.Equal(t, "https://111111111111.dkr.ecr.ap-northeast-1.amazonaws.com/v2/strix/blobs/sha256:b5b2b2c507a0944348e0303114d8d93aaaa081732b86451d9bce1f432a537bc7", req2.URL.String())
			assert.Equal(t, "Basic fake_auth_token", req2.Header.Get("Authorization"))
			assert.Equal(t, "application/vnd.docker.distribution.manifest.v2+json", req2.Header.Get("Accept"))
		})
	})
}

func newControllerForTrivyScanImageTest(t *testing.T) (*controller.Controller, *mockSet) {
	mock := &mockSet{}

	ctrl := &controller.Controller{
		Config: controller.Config{
			AwsRegion:       "us-east-0",
			TableName:       "trivy-scan-test",
			S3Region:        "ap-northeast-0",
			S3Bucket:        "example-bucket",
			S3Prefix:        "testing/",
			ScanQueueURL:    "https://sqs.us-east-2.amazonaws.com/123456789012/scan-queue",
			InspectQueueURL: "https://sqs.us-east-2.amazonaws.com/123456789012/inspect-queue",
		},
	}

	ctrl.InjectAdaptors(controller.Adaptors{
		NewS3: func(region string) (interfaces.S3Client, error) {
			if mock.s3.regions == nil {
				mock.s3.regions = make(map[string]struct{})
			}
			mock.s3.regions[region] = struct{}{}
			return &mock.s3, nil
		},
		NewSQS:   func(region string) (interfaces.SQSClient, error) { return &mock.sqs, nil },
		NewECR:   func(region string) (interfaces.ECRClient, error) { return &mock.ecr, nil },
		MkdirAll: func(path string, perm os.FileMode) error { return nil },
		Exec:     func(s1 string, s2 ...string) ([]byte, error) { return nil, nil },
		Create: func(s string) (io.WriteCloser, error) {
			buf := &bufCloser{}
			mock.buffers = append(mock.buffers, buf)
			return buf, nil
		},

		NewDBClient: func(region, tableName string) (interfaces.DBClient, error) {
			var err error
			mock.dbClient, err = db.NewDynamoClientLocal(region, tableName)
			return mock.dbClient, err
		},
		ReadFile: func(filename string) ([]byte, error) {
			return ioutil.ReadFile(path.Join("..", "testdata", "trivy_output_sample1.json"))
		},
		TempFile: interfaces.DefaultTempFileFunc,
		HTTP:     &mock.http,
	})

	dbClient := ctrl.DB()
	t.Logf("dynamo table name: %s", dbClient.(*db.DynamoClient).TableName())

	t.Cleanup(func() {
		if !t.Failed() {
			if err := ctrl.DB().Close(); err != nil {
				t.Fatal("Can not close DB: ", err)
			}
		}
	})

	// ECR
	mock.ecr.getTokenOutput = []*ecr.GetAuthorizationTokenOutput{
		{
			AuthorizationData: []*ecr.AuthorizationData{
				{
					AuthorizationToken: aws.String("fake_auth_token"),
				},
			},
		},
	}

	manifestData := `{
	"schemaVersion": 2,
	"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
	"config": {
		"mediaType": "application/vnd.docker.container.image.v1+json",
		"size": 7023,
		"digest": "sha256:b5b2b2c507a0944348e0303114d8d93aaaa081732b86451d9bce1f432a537bc7"
	},
	"layers": [
		{
			"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
			"size": 32654,
			"digest": "sha256:e692418e4cbaf90ca69d05a66403747baa33ee08806650b51fab815ad7fc331f"
		},
		{
			"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
			"size": 16724,
			"digest": "sha256:3c3a4604a545cdc127456d94e421cd355bca5b528f4a9c1905b15da2eb4a4c6b"
		},
		{
			"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
			"size": 73109,
			"digest": "sha256:ec4b8955958665577945c89419d1af06b5f7636b4ac3da7f12184802ad867736"
		}
	]
}`

	// Requires only container_config.Env
	configData := `{
	"container": "0fba8b9abff35ecac0aab193b6222e74faf48d19829e7xxxxxxxxxxxxxxx",
	"container_config": {
		"Env": [
			"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
			"DEBIAN_FRONTEND=noninteractive",
			"LANG=C.UTF-8"
		]
	},
	"created": "2021-03-11T01:19:55.595142316Z",
	"os": "linux"
}
`
	mock.http.responses = []*http.Response{
		// for GetImageManifest
		{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(manifestData)),
		},
		// for GetImageEnv
		{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(configData)),
		},
	}

	return ctrl, mock
}
