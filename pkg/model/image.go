package model

import (
	"strings"

	"github.com/m-mizutani/golambda"
)

// TaggedImage is docker image identifier
type TaggedImage struct {
	Registry string
	Repo     string
	Tag      string
}

type Image struct {
	Registry string
	Repo     string
	Digest   string
}

type ImageMeta struct {
	LayerDigests []string
	Env          []string
}

// RegistryRepoTag returns "{registry}/{repo}:{tag}"
func (x *TaggedImage) RegistryRepoTag() string {
	return x.Registry + "/" + x.Repo + ":" + x.Tag
}

// RegistryRepoDigest returns "{registry}/{repo}:{digest}"
func (x *Image) RegistryRepoDigest() string {
	return x.Registry + "/" + x.Repo + ":" + x.Digest
}

// ParseRepositoryURI parses URL like 11111111.dkr.ecr.ap-northeast-1.amazonaws.com/some-image and returns filled Image object.
func ParseRepositoryURI(uri string) (*TaggedImage, error) {
	parts := strings.Split(uri, "/")
	if len(parts) != 2 {
		return nil, golambda.NewError("Invalid repository URI format, number of slash charactor is not one").With("uri", uri)
	}

	return &TaggedImage{
		Registry: parts[0],
		Repo:     parts[1],
	}, nil
}

type ImageLayerIndex struct {
	Image
	LayerDigest string
}
