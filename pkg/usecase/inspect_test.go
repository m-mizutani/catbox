package usecase_test

import (
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
		ctrl, _ := newControllerForInspectTest(t)
		req := &model.InspectRequestMessage{
			ReportID: "no-such-report",
		}

		assert.ErrorIs(t, usecase.InspectScanReport(ctrl, req), usecase.ErrReportNotFound)
	})

	t.Run("Normal case", func(t *testing.T) {
		ctrl, mock := newControllerForInspectTest(t)
		req := &model.InspectRequestMessage{
			ReportID: "report-id-1",
		}

		// Run usecase
		assert.NoError(t, usecase.InspectScanReport(ctrl, req))
	})
}

func newControllerForInspectTest(t *testing.T) (*controller.Controller, *mockSet) {
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

	// -------------------------------
	// insert base data
	report := &model.ScanReport{
		ReportID:  "report-id-1", // Usually ReportID should not be filled, but allow it for testing
		StatusSeq: 10,
		Image: model.Image{
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

	// Trivy scan result
	trivyResultData, err := ioutil.ReadFile("../testdata/trivy_output_sample1.json")
	require.NoError(t, err)
	mock.s3.saveObject("report-bucket", "path/to/my/report.json.gz", trivyResultData)

	// Save report data
	require.NoError(t, mock.dbClient.PutScanReport(report))

	return ctrl, mock
}
