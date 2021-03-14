package db_test

import (
	"testing"

	"github.com/m-mizutani/catbox/pkg/interfaces"
	"github.com/m-mizutani/catbox/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVulnInfo(t *testing.T) {
	t.Run("get existing vulnInfo", func(t *testing.T) {
		client, inserted := setupVulnInfoTest(t)
		vulnSet, err := client.GetVulnInfoBatch([]string{"CVE-2001-0001", "CVE-2999-9999"})
		require.NoError(t, err)
		require.Equal(t, 1, len(vulnSet))
		assert.Equal(t, inserted[0], vulnSet[0])
	})

	t.Run("put additional vulnInfo", func(t *testing.T) {
		client, inserted := setupVulnInfoTest(t)
		newOne := *&model.VulnInfo{
			ID:         "CVE-1983-0420",
			Title:      "death",
			DetectedAt: 1,
		}
		newSet, err := client.PutVulnInfoBatch([]*model.VulnInfo{inserted[0], &newOne})
		require.NoError(t, err)
		assert.Equal(t, 1, len(newSet)) // returned only new one
		assert.Contains(t, newSet, &newOne)
	})
}

func setupVulnInfoTest(t *testing.T) (interfaces.DBClient, []*model.VulnInfo) {
	client := newTestTable(t)

	vulnInfoSet := []*model.VulnInfo{
		{
			ID:         "CVE-2001-0001",
			Title:      "blue",
			DetectedAt: 1000,
			References: []string{},
		},
		{
			ID:         "CVE-2001-0002",
			Title:      "orange",
			DetectedAt: 1000,
			References: []string{},
		},
		{
			ID:         "CVE-2001-0003",
			Title:      "red",
			DetectedAt: 1000,
			References: []string{},
		},
	}
	newVuln, err := client.PutVulnInfoBatch(vulnInfoSet)
	require.NoError(t, err)
	require.Equal(t, 3, len(newVuln))
	assert.Contains(t, newVuln, vulnInfoSet[0])
	assert.Contains(t, newVuln, vulnInfoSet[1])
	assert.Contains(t, newVuln, vulnInfoSet[2])

	return client, vulnInfoSet
}
