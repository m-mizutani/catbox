package db

import (
	"github.com/m-mizutani/catbox/pkg/model"
	"github.com/m-mizutani/golambda"
)

func imageLayerIndexPK(layerDigest string) string {
	return "layer_digest:" + layerDigest
}

// PutImageLayerDigest inserts layerDigest
func (x *DynamoClient) PutImageLayerDigest(index *model.ImageLayerIndex) error {
	record := dynamoRecord{
		PK:  imageLayerIndexPK(index.LayerDigest),
		SK:  index.RegistryRepoDigest(),
		Doc: index,
	}

	if err := x.table.Put(record).Run(); err != nil {
		return golambda.WrapError(err, "PutImageLayerDigest").With("index", index)
	}
	return nil
}

// LookupImageLayerDigest returns image that has specified layer digest. It returns nil if no index is found
func (x *DynamoClient) LookupImageLayerDigest(digest string) ([]*model.ImageLayerIndex, error) {
	pk := imageLayerIndexPK(digest)
	var records []dynamoRecord
	if err := x.table.Get("pk", pk).All(&records); err != nil {
		return nil, golambda.WrapError(err, "LookupImageLayerDigests").With("digest", digest)
	}

	resp := make([]*model.ImageLayerIndex, len(records))
	for i := range records {
		if err := records[i].Unmarshal(&resp[i]); err != nil {
			return nil, err
		}
	}

	return resp, nil
}
