package db_test

import (
	"testing"

	"github.com/m-mizutani/catbox/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
