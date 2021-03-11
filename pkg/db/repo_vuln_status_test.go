package db_test

import (
	"testing"

	"github.com/m-mizutani/catbox/pkg/interfaces"
	"github.com/m-mizutani/catbox/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRepoVulnStatusTemplate() *model.RepoVulnStatus {
	return &model.RepoVulnStatus{
		TaggedImage: model.TaggedImage{
			Registry: "1111111111.dkr.ecr.ap-northeast-1.amazonaws.com",
			Repo:     "test-image",
			Tag:      "good-tag",
			Digest:   "12345678",
		},

		RepoVulnEntry: model.RepoVulnEntry{
			VulnID:    "CVE-2001-1234",
			VulnType:  model.VulnPkg,
			PkgSource: "/tmp/Gemfile.lock",
			PkgName:   "nanika",
		},
		UpdatedAt:  12345,
		Status:     model.VulnStatusNew,
		DetectedBy: model.ScannerTrivy,
		StatusSeq:  5,

		PkgType:             "bundler",
		PkgInstalledVersion: "1.1",
		PkgFixedVersion:     "1.2",
	}
}

func insertBaseRepoVulnStatus(t *testing.T, client interfaces.DBClient) []*model.RepoVulnStatus {
	s1 := newRepoVulnStatusTemplate()
	s2 := newRepoVulnStatusTemplate()
	s2.VulnID = "CVE-2020-0225"

	s3 := newRepoVulnStatusTemplate() // Must not be found by s1 image
	s3.Repo = "other-repo"

	r1, err1 := client.CreateRepoVulnStatus(s1)
	require.NoError(t, err1)
	assert.True(t, r1)
	r2, err2 := client.CreateRepoVulnStatus(s2)
	require.NoError(t, err2)
	assert.True(t, r2)
	r3, err3 := client.CreateRepoVulnStatus(s3)
	require.NoError(t, err3)
	assert.True(t, r3)

	return []*model.RepoVulnStatus{s1, s2, s3}
}

func TestRepoVulnStatus(t *testing.T) {
	t.Run("Put RepoVulnStatus items", func(t *testing.T) {
		client := newTestTable(t)

		s := insertBaseRepoVulnStatus(t, client)

		t.Run("And get them by repo", func(t *testing.T) {
			img := model.TaggedImage{
				Registry: "1111111111.dkr.ecr.ap-northeast-1.amazonaws.com",
				Repo:     "test-image",
				Tag:      "good-tag",
			}
			resp, err := client.GetRepoVulnStatusByRepo(&img)
			require.NoError(t, err)
			require.Equal(t, 2, len(resp))
			assert.Contains(t, resp, s[0])
			assert.Contains(t, resp, s[1])
		})

		t.Run("And get them by vulnID", func(t *testing.T) {
			resp, err := client.GetRepoVulnStatusByVulnID("CVE-2001-1234")
			require.NoError(t, err)
			require.Equal(t, 2, len(resp))
			assert.Contains(t, resp, s[0])
			assert.Contains(t, resp, s[2])
		})
	})

	t.Run("Update RepoStatus", func(t *testing.T) {
		genBaseChangeLog := func() *model.RepoVulnChangeLog {
			return &model.RepoVulnChangeLog{
				TaggedImage: model.TaggedImage{
					Registry: "1111111111.dkr.ecr.ap-northeast-1.amazonaws.com",
					Repo:     "test-image",
					Tag:      "good-tag",
				},
				RepoVulnEntry: model.RepoVulnEntry{
					VulnID:    "CVE-2001-1234",
					VulnType:  model.VulnPkg,
					PkgSource: "/tmp/Gemfile.lock",
					PkgName:   "nanika",
				},
				Status:    model.VulnStatusFixed,
				UpdatedAt: 1234,
				StatusSeq: 10,
			}
		}

		t.Run("Normal case", func(t *testing.T) {
			client := newTestTable(t)
			_ = insertBaseRepoVulnStatus(t, client)

			t.Run("Normal case", func(t *testing.T) {
				c := genBaseChangeLog()
				before1, err := client.GetRepoVulnChangeLogs(&c.TaggedImage)
				require.NoError(t, err)
				assert.Equal(t, 2, len(before1))
				assert.Equal(t, model.VulnStatusNew, before1[0].Status)
				assert.Equal(t, model.VulnStatusNew, before1[1].Status)

				before2, err := client.GetRepoVulnEntryChangeLogs(&c.TaggedImage, &c.RepoVulnEntry)
				require.NoError(t, err)
				assert.Equal(t, 1, len(before2))
				assert.Equal(t, model.VulnStatusNew, before2[0].Status)
				assert.Equal(t, "CVE-2001-1234", before2[0].VulnID)

				// Run update
				updated, err := client.UpdateRepoVulnStatus(c)
				require.NoError(t, err)
				assert.True(t, updated)

				// Check results of update
				after1, err := client.GetRepoVulnChangeLogs(&c.TaggedImage)
				require.NoError(t, err)
				assert.Equal(t, 3, len(after1))

				after2, err := client.GetRepoVulnEntryChangeLogs(&c.TaggedImage, &c.RepoVulnEntry)
				require.NoError(t, err)
				assert.Equal(t, 2, len(after2))
				assert.Equal(t, model.VulnStatusNew, after2[0].Status)
				assert.Equal(t, "CVE-2001-1234", after2[0].VulnID)
				assert.Equal(t, model.VulnStatusFixed, after2[1].Status)
				assert.Equal(t, "CVE-2001-1234", after2[1].VulnID)
			})
		})

		t.Run("entry is not found", func(t *testing.T) {
			client := newTestTable(t)
			_ = insertBaseRepoVulnStatus(t, client)

			t.Run("Non existing registry", func(t *testing.T) {
				c := genBaseChangeLog()
				img := &model.TaggedImage{
					Registry: "1111111111.dkr.ecr.ap-northeast-1.amazonaws.com",
					Repo:     "test-image",
					Tag:      "good-tag",
				}

				before, err := client.GetRepoVulnEntryChangeLogs(img, &c.RepoVulnEntry)
				require.NoError(t, err)
				assert.Equal(t, 1, len(before))

				c.Registry = "non-exist-registry"
				updated, err := client.UpdateRepoVulnStatus(c)
				assert.NoError(t, err)
				assert.False(t, updated)

				after, err := client.GetRepoVulnEntryChangeLogs(img, &c.RepoVulnEntry)
				require.NoError(t, err)
				assert.Equal(t, 1, len(after))
			})

			t.Run("Non existing repo", func(t *testing.T) {
				c := genBaseChangeLog()
				c.Repo = "non-exist-repo"
				updated, err := client.UpdateRepoVulnStatus(c)
				assert.NoError(t, err)
				assert.False(t, updated)
			})

			t.Run("Non existing tag", func(t *testing.T) {
				c := genBaseChangeLog()
				c.Tag = "non-exist-tag"
				updated, err := client.UpdateRepoVulnStatus(c)
				assert.NoError(t, err)
				assert.False(t, updated)
			})

			t.Run("Non existing vulnID", func(t *testing.T) {
				c := genBaseChangeLog()
				c.VulnID = "?"
				updated, err := client.UpdateRepoVulnStatus(c)
				assert.NoError(t, err)
				assert.False(t, updated)
			})

			t.Run("Non existing pkg source", func(t *testing.T) {
				c := genBaseChangeLog()
				c.PkgSource = "?"
				updated, err := client.UpdateRepoVulnStatus(c)
				assert.NoError(t, err)
				assert.False(t, updated)
			})

			t.Run("Non existing pkg name", func(t *testing.T) {
				c := genBaseChangeLog()
				c.PkgSource = "non-exist-pkg"
				updated, err := client.UpdateRepoVulnStatus(c)
				assert.NoError(t, err)
				assert.False(t, updated)
			})
		})

		t.Run("condition is not matched", func(t *testing.T) {
			client := newTestTable(t)
			_ = insertBaseRepoVulnStatus(t, client)

			t.Run("seq is same", func(t *testing.T) {
				c := genBaseChangeLog()
				c.StatusSeq = 5
				updated, err := client.UpdateRepoVulnStatus(c)
				assert.NoError(t, err)
				assert.False(t, updated)
			})

			t.Run("seq is old", func(t *testing.T) {
				c := genBaseChangeLog()
				c.StatusSeq = 4
				updated, err := client.UpdateRepoVulnStatus(c)
				assert.NoError(t, err)
				assert.False(t, updated)
			})

			t.Run("status is not changed", func(t *testing.T) {
				c := genBaseChangeLog()
				c.Status = model.VulnStatusNew
				updated, err := client.UpdateRepoVulnStatus(c)
				assert.NoError(t, err)
				assert.False(t, updated)
			})
		})
	})
}
