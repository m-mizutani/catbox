package db_test

import (
	"testing"

	"github.com/m-mizutani/catbox/pkg/db"
	"github.com/m-mizutani/catbox/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestTable(t *testing.T) *db.DynamoClient {
	tableName := "dynamodb-test"

	client, err := db.NewDynamoClientLocal("ap-northeast-1", tableName)
	if err != nil {
		panic("Failed to use local DynamoDB: " + err.Error())
	}

	t.Log("Created table name: ", client.TableName())
	return client
}

func deleteTestTable(t *testing.T, client *db.DynamoClient) {
	if t.Failed() {
		return // Failed test table is not deleted
	}

	if err := client.Close(); err != nil {
		panic("Failed to delete test table: " + err.Error())
	}
}

func newRepoVulnStatusTemplate() *model.RepoVulnStatus {
	return &model.RepoVulnStatus{
		Image: model.Image{
			Registry:     "1111111111.dkr.ecr.ap-northeast-1.amazonaws.com",
			Repo:         "test-image",
			Tag:          "good-tag",
			Digest:       "12345678",
			LayerDigests: make([]string, 0),
			Env:          make([]string, 0),
		},

		VulnID:    "CVE-2001-1234",
		UpdatedAt: 12345,
		Status:    model.VulnStatusNew,
		ScanType:  model.ScanTypeTrivy,

		PkgSource:        "/tmp/Gemfile.lock",
		PkgName:          "nanika",
		PkgType:          "bundler",
		InstalledVersion: "1.1",
		FixedVersion:     "1,2",
	}
}

func newScanReporTemplate() *model.ScanReport {
	return &model.ScanReport{
		Image: model.Image{
			Registry:     "1111111111.dkr.ecr.ap-northeast-1.amazonaws.com",
			Repo:         "star",
			Tag:          "main",
			Digest:       "12345678",
			LayerDigests: make([]string, 0),
			Env:          make([]string, 0),
		},
		ScanType:    model.ScanTypeTrivy,
		RequestedAt: 123456,
		RequestedBy: "ecr.PutImage",
		InvokedAt:   234567,
		ScannedAt:   234569,

		S3Region: "us-east-0",
		S3Bucket: "example-bucket",
		S3Key:    "hoge/moge.json.gz",
	}
}

func TestRepoVulnStatus(t *testing.T) {
	t.Run("Put RepoVulnStatus items", func(t *testing.T) {
		client := newTestTable(t)
		defer deleteTestTable(t, client)

		status1 := newRepoVulnStatusTemplate()
		status2 := newRepoVulnStatusTemplate()
		status2.VulnID = "CVE-2020-0225"

		otherRepo := newRepoVulnStatusTemplate() // Must not be found by status1 image
		otherRepo.Repo = "other-repo"

		require.NoError(t, client.PutRepoVulnStatusBatch([]*model.RepoVulnStatus{
			status1, status2, otherRepo,
		}))

		t.Run("And get them by repo", func(t *testing.T) {
			resp, err := client.GetRepoVulnStatusByRepo("1111111111.dkr.ecr.ap-northeast-1.amazonaws.com", "test-image", "good-tag")
			require.NoError(t, err)
			require.Equal(t, 2, len(resp))
			assert.Contains(t, resp, status1)
			assert.Contains(t, resp, status2)
		})

		t.Run("And get them by vulnID", func(t *testing.T) {
			resp, err := client.GetRepoVulnStatusByVulnID("CVE-2001-1234")
			require.NoError(t, err)
			require.Equal(t, 2, len(resp))
			assert.Contains(t, resp, status1)
			assert.Contains(t, resp, otherRepo)
		})
	})

	t.Run("Put new RepoVulnStatus after putting old one and get them", func(t *testing.T) {
		client := newTestTable(t)
		defer deleteTestTable(t, client)

		status1 := newRepoVulnStatusTemplate()
		require.NoError(t, client.PutRepoVulnStatusBatch([]*model.RepoVulnStatus{status1}))

		status2 := newRepoVulnStatusTemplate()
		status2.UpdatedAt += 10
		status2.Status = model.VulnStatusFixed
		require.NoError(t, client.PutRepoVulnStatusBatch([]*model.RepoVulnStatus{status2}))

		resp, err := client.GetRepoVulnStatusByRepo("1111111111.dkr.ecr.ap-northeast-1.amazonaws.com", "test-image", "good-tag")
		require.NoError(t, err)
		require.Equal(t, 2, len(resp))
		assert.Contains(t, resp, status1)
		assert.Contains(t, resp, status2)
	})
}

func TestScanReport(t *testing.T) {
	t.Run("Put and get reports", func(t *testing.T) {
		client := newTestTable(t)
		defer deleteTestTable(t, client)

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

func TestImageLayerDigests(t *testing.T) {
	t.Run("Put and lookup layer digests", func(t *testing.T) {
		client := newTestTable(t)
		defer deleteTestTable(t, client)

		idx1 := &model.ImageLayerIndex{
			Image: model.Image{
				Registry:     "1111111111.dkr.ecr.ap-northeast-1.amazonaws.com",
				Repo:         "blue",
				Digest:       "abc123",
				LayerDigests: make([]string, 0),
				Env:          make([]string, 0),
			},
			LayerDigest: "caffee",
		}
		idx2 := &model.ImageLayerIndex{
			Image: model.Image{
				Registry:     "1111111111.dkr.ecr.ap-northeast-1.amazonaws.com",
				Repo:         "orange",
				Digest:       "321bca",
				LayerDigests: make([]string, 0),
				Env:          make([]string, 0),
			},
			LayerDigest: "beef00",
		}
		idx3 := &model.ImageLayerIndex{
			Image: model.Image{
				Registry:     "1111111111.dkr.ecr.ap-northeast-1.amazonaws.com",
				Repo:         "five",
				Digest:       "112233",
				LayerDigests: make([]string, 0),
				Env:          make([]string, 0),
			},
			LayerDigest: "xxxxxx",
		}

		require.NoError(t, client.PutImageLayerDigest(idx1))
		require.NoError(t, client.PutImageLayerDigest(idx2))
		require.NoError(t, client.PutImageLayerDigest(idx3))

		layer1, err := client.LookupImageLayerDigest("caffee")
		require.NoError(t, err)
		require.Equal(t, 1, len(layer1))
		assert.Equal(t, layer1[0], idx1)

		layer2, err := client.LookupImageLayerDigest("beef00")
		require.NoError(t, err)
		require.Equal(t, 1, len(layer2))
		assert.Equal(t, layer2[0], idx2)

		layer3, err := client.LookupImageLayerDigest("?")
		require.NoError(t, err)
		require.Equal(t, 0, len(layer3))
	})
}
