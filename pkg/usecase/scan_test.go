package usecase_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/google/uuid"
	"github.com/m-mizutani/catbox/pkg/controller"
	"github.com/m-mizutani/catbox/pkg/db"
	"github.com/m-mizutani/catbox/pkg/interfaces"
	"github.com/m-mizutani/catbox/pkg/model"
	"github.com/m-mizutani/catbox/pkg/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	mockS3 struct {
		regions         map[string]struct{}
		getSeq          int
		getObjectInput  []*s3.GetObjectInput
		getObjectOutput []*s3.GetObjectOutput

		putObjectInput []*s3.PutObjectInput
	}
	mockSQS struct {
		input []*sqs.SendMessageInput
	}
	bufCloser struct{ bytes.Buffer }

	mockSet struct {
		s3       mockS3
		sqs      mockSQS
		buffers  []*bufCloser
		dbClient interfaces.DBClient
	}
)

func (x *mockS3) GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	x.getObjectInput = append(x.getObjectInput, input)
	x.getSeq++
	return x.getObjectOutput[x.getSeq-1], nil
}

func (x *mockS3) PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	x.putObjectInput = append(x.putObjectInput, input)
	return &s3.PutObjectOutput{}, nil
}

func (x *mockSQS) SendMessage(input *sqs.SendMessageInput) (*sqs.SendMessageOutput, error) {
	x.input = append(x.input, input)
	return &sqs.SendMessageOutput{}, nil
}

func (x *bufCloser) Close() error { return nil }

func teardownController(t *testing.T, ctrl *controller.Controller) {
	if !t.Failed() {
		ctrl.DB().Close()
	}
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

	return ctrl, mock
}

func TestTrivyScanImage(t *testing.T) {
	t.Run("normal case", func(t *testing.T) {
		ctrl, mock := newControllerForTrivyScanImageTest(t)

		mock.s3.getObjectOutput = []*s3.GetObjectOutput{
			{Body: ioutil.NopCloser(strings.NewReader("test"))},
			{Body: ioutil.NopCloser(strings.NewReader("test"))},
		}
		req := &model.ScanRequestMessage{
			RequestID:   uuid.New().String(),
			RequestedBy: "ecr.PutImage",
			RequestedAt: time.Now().UTC(),
			Target: model.Image{
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
		})

		t.Run("report is saved on S3", func(t *testing.T) {
			require.Equal(t, 1, len(mock.s3.putObjectInput))
			assert.Contains(t, mock.s3.regions, "ap-northeast-0")
			assert.Equal(t, "example-bucket", *mock.s3.putObjectInput[0].Bucket)
			assert.Equal(t, "testing/path/to/prefix/trivy.json.gz", *mock.s3.putObjectInput[0].Key)
		})
	})
}
