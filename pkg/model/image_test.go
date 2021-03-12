package model_test

import (
	"testing"

	"github.com/m-mizutani/catbox/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestImage(t *testing.T) {
	tagImg := model.TaggedImage{
		Registry: "test.registry.com",
		Repo:     "blue",
		Tag:      "magic",
	}
	assert.Equal(t, "test.registry.com/blue:magic", tagImg.RegistryRepoTag())

	img := model.Image{
		Registry: "test.registry.com",
		Repo:     "blue",
		Digest:   "beefcafe",
	}
	assert.Equal(t, "test.registry.com/blue:beefcafe", img.RegistryRepoDigest())
}
