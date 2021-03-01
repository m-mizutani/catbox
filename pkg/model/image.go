package model

// Image is docker image identifier
type Image struct {
	Registry string
	Repo     string
	Tag      string
	Digest   string `json:",omitempty"`
}

type ImageMeta struct {
	LayerDigests []string
	Env          []string
}

// RegistryRepoTag returns "{registry}/{repo}:{tag}"
func (x *Image) RegistryRepoTag() string {
	return x.Registry + "/" + x.Repo + ":" + x.Tag
}

// RegistryRepoDigest returns "{registry}/{repo}:{digest}"
func (x *Image) RegistryRepoDigest() string {
	return x.Registry + "/" + x.Repo + ":" + x.Digest
}

type ImageLayerIndex struct {
	Image
	LayerDigest string
}
