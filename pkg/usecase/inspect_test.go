package usecase_test

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/m-mizutani/catbox/pkg/controller"
	"github.com/m-mizutani/catbox/pkg/db"
	"github.com/m-mizutani/catbox/pkg/interfaces"
	"github.com/m-mizutani/catbox/pkg/model"
	"github.com/m-mizutani/catbox/pkg/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInspect(t *testing.T) {
	t.Run("Error with invalid report ID", func(t *testing.T) {
		ctrl, _ := setupControllerForInspectTest(t)
		req := &model.InspectRequestMessage{
			ReportID: "no-such-report",
		}

		assert.ErrorIs(t, usecase.InspectScanReport(ctrl, req), usecase.ErrReportNotFound)
	})

	t.Run("Normal case", func(t *testing.T) {
		ctrl, mock := setupControllerForInspectTest(t)
		req := &model.InspectRequestMessage{
			ReportID: "report-id-1",
		}

		// Run usecase
		assert.NoError(t, usecase.InspectScanReport(ctrl, req))

		// Check outputs
		t.Run("Published change message to SNS", func(t *testing.T) {
			require.Equal(t, 2, len(mock.sns.input))

			var msg1, msg2 model.ChangeMessage
			// msg1: new vulnerabilities notification
			require.NoError(t, json.Unmarshal([]byte(*mock.sns.input[0].Message), &msg1))
			require.Equal(t, 4, len(msg1.NewVuln))
			require.Equal(t, 0, len(msg1.UpdatedStatus))

			// msg2: new RepoVulnStat notification
			require.NoError(t, json.Unmarshal([]byte(*mock.sns.input[1].Message), &msg2))
			assert.Equal(t, 0, len(msg2.NewVuln))
			require.Equal(t, 5, len(msg2.UpdatedStatus))
			vulnIDs := make([]string, len(msg2.UpdatedStatus))
			for i := range vulnIDs {
				vulnIDs[i] = msg2.UpdatedStatus[i].VulnID
			}
			assert.Contains(t, vulnIDs, "CVE-2020-27350")
			assert.Contains(t, vulnIDs, "CVE-2017-15131")
			assert.Contains(t, vulnIDs, "CVE-2020-7662")
			assert.Contains(t, vulnIDs, "GHSA-p9pc-299p-vxgp")
		})

		t.Run("Added new RepoVulnStatus", func(t *testing.T) {
			stats, err := mock.dbClient.GetRepoVulnStatusByRepo(&model.TaggedImage{
				Registry: "111111111111.dkr.ecr.ap-northeast-1.amazonaws.com",
				Repo:     "strix",
				Tag:      "latest",
			})
			require.NoError(t, err)
			assert.Equal(t, 5, len(stats))

			vulnIDs := make([]string, len(stats))
			for i := range vulnIDs {
				vulnIDs[i] = stats[i].VulnID
			}
			assert.Contains(t, vulnIDs, "CVE-2020-27350")
			assert.Contains(t, vulnIDs, "CVE-2017-15131")
			assert.Contains(t, vulnIDs, "CVE-2020-7662")
			assert.Contains(t, vulnIDs, "GHSA-p9pc-299p-vxgp")

			var GHSAp9pc299pvxgpVersions []string
			for _, s := range stats {
				if s.VulnID == "GHSA-p9pc-299p-vxgp" {
					GHSAp9pc299pvxgpVersions = append(GHSAp9pc299pvxgpVersions, s.PkgVersion)
				}
			}
			assert.Equal(t, 2, len(GHSAp9pc299pvxgpVersions))
			assert.Contains(t, GHSAp9pc299pvxgpVersions, "5.0.0")
			assert.Contains(t, GHSAp9pc299pvxgpVersions, "13.1.1")
		})
	})

	t.Run("Check idempotence", func(t *testing.T) {
		ctrl, mock := setupControllerForInspectTest(t)
		req := &model.InspectRequestMessage{
			ReportID: "report-id-1",
		}

		// Run usecase twice
		assert.NoError(t, usecase.InspectScanReport(ctrl, req))
		assert.NoError(t, usecase.InspectScanReport(ctrl, req))

		// No additional vulnInfo and RepoVulnStatus is not updated
		require.Equal(t, 2, len(mock.sns.input))

		// Number of RepoVulnStatus is still 4
		stats, err := mock.dbClient.GetRepoVulnStatusByRepo(&model.TaggedImage{
			Registry: "111111111111.dkr.ecr.ap-northeast-1.amazonaws.com",
			Repo:     "strix",
			Tag:      "latest",
		})
		require.NoError(t, err)
		assert.Equal(t, 5, len(stats))
	})
}

func setupControllerForInspectTest(t *testing.T) (*controller.Controller, *mockSet) {
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
			ChangeTopicARN:  "arn:aws:sns:us-east-1:111122223333:my-topic",
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
		NewSNS:   func(region string) (interfaces.SNSClient, error) { return &mock.sns, nil },
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

	// -------------------------------
	// insert base data

	// Scan report
	report := &model.ScanReport{
		ReportID:  "report-id-1", // Usually ReportID should not be filled, but allow it for testing
		StatusSeq: 10,
		TaggedImage: model.TaggedImage{
			Registry: "111111111111.dkr.ecr.ap-northeast-1.amazonaws.com",
			Repo:     "strix",
			Tag:      "latest",
		},
		ImageMeta:   model.ImageMeta{},
		ScannedBy:   model.ScannerTrivy,
		RequestedAt: 1234,
		InvokedAt:   2345,
		ScannedAt:   3456,
		RequestedBy: "ecr.PushImage",
		OutputTo: model.S3Path{
			Region: "ap-northeast-0",
			Bucket: "report-bucket", // basically Config.S3Bucket and OutputTo.Bucket should be same. However other bucket is also acceptable for changing main S3 bucket safely.
			Key:    "path/to/my/report.json.gz",
		},
	}
	require.NoError(t, mock.dbClient.PutScanReport(report))

	// Trivy scan result
	trivyResultData, err := ioutil.ReadFile("../testdata/trivy_output_sample1.json")
	require.NoError(t, err)
	mock.s3.saveObject("report-bucket", "path/to/my/report.json.gz", trivyResultData)

	return ctrl, mock
}
