package db_test

import (
	"testing"

	"github.com/m-mizutani/catbox/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
