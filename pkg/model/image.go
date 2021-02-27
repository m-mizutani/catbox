package model

// Image is docker image identifier
type Image struct {
	Registry     string   `dynamo:"registry" json:"registry"`
	Repo         string   `dynamo:"repo" json:"repo"`
	Tag          string   `dynamo:"tag" json:"tag"`
	Digest       string   `dynamo:"digest,omitempty" json:"digest,omitempty"`
	LayerDigests []string `dynamo:"layer_digests,omitempty" json:"layer_digests,omitempty"`
	Env          []string `dynamo:"env,omitempty" json:"env,omitempty"`
}

// RegistryRepoTag returns "{registry}/{repo}:{tag}"
func (x *Image) RegistryRepoTag() string {
	return x.Registry + "/" + x.Repo + ":" + x.Tag
}

type ImageLayerIndex struct {
	DBBaseRecord
	Image
	LayerDigest string `dynamo:"layer_digest" json:"layer_digest"`
}
