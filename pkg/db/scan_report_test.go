package db_test

import (
	"testing"

	"github.com/m-mizutani/catbox/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newScanReporTemplate() *model.ScanReport {
	return &model.ScanReport{
		TaggedImage: model.TaggedImage{
			Registry: "1111111111.dkr.ecr.ap-northeast-1.amazonaws.com",
			Repo:     "star",
			Tag:      "main",
			Digest:   "12345678",
		},
		ImageMeta: model.ImageMeta{
			LayerDigests: make([]string, 0),
			Env:          make([]string, 0),
		},
		ScannedBy:   model.ScannerTrivy,
		RequestedAt: 123456,
		RequestedBy: "ecr.PutImage",
		InvokedAt:   234567,
		ScannedAt:   234569,

		OutputTo: model.S3Path{
			Region: "us-east-0",
			Bucket: "example-bucket",
			Key:    "hoge/moge.json.gz",
		},
	}
}

func TestScanReport(t *testing.T) {
	t.Run("Put and get reports", func(t *testing.T) {
		client := newTestTable(t)

		report1 := newScanReporTemplate()
		require.NoError(t, client.PutScanReport(report1))

		// same with report1, but different timestamp
		report2 := newScanReporTemplate()
		report2.ScannedAt += 1
		require.NoError(t, client.PutScanReport(report2))

		// other repository with report1
		report3 := newScanReporTemplate()
		report3.ScannedAt += 2
		report3.Repo = "moon"
		require.NoError(t, client.PutScanReport(report3))

		t.Run("ReportID is assigned automatically when putting and unique", func(t *testing.T) {
			require.NotEmpty(t, report1.ReportID)
			require.NotEmpty(t, report2.ReportID)
			assert.NotEqual(t, report1.ReportID, report2.ReportID)
		})

		t.Run("Lookup with report ID", func(t *testing.T) {
			resp, err := client.GetScanReportByID(report1.ReportID)
			require.NoError(t, err)
			assert.Equal(t, resp, report1)
		})

		t.Run("Not found with invalid report ID", func(t *testing.T) {
			resp, err := client.GetScanReportByID("?")
			require.NoError(t, err)
			assert.Nil(t, resp)
		})

		t.Run("Lookup with repo", func(t *testing.T) {
			resp, err := client.GetLatestScanReportsByRepo("1111111111.dkr.ecr.ap-northeast-1.amazonaws.com", "star", "main")
			require.NoError(t, err)
			require.Equal(t, report2, resp)
		})

		t.Run("Not found with invalid repo", func(t *testing.T) {
			resp, err := client.GetLatestScanReportsByRepo("1111111111.dkr.ecr.ap-northeast-1.amazonaws.com", "not_found", "main")
			require.NoError(t, err)
			require.Nil(t, resp)
		})
	})
}
