package model

// Image is docker image identifier
type Image struct {
	Registry     string
	Repo         string
	Tag          string
	Digest       string   `json:",omitempty"`
	LayerDigests []string `json:",omitempty"`
	Env          []string `json:",omitempty"`
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
