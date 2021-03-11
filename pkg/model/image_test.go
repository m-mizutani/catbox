package model_test

import (
	"testing"

	"github.com/m-mizutani/catbox/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestImage(t *testing.T) {
	img := model.TaggedImage{
		Registry: "test.registry.com",
		Repo:     "blue",
		Tag:      "magic",
		Digest:   "beefcafe",
	}

	assert.Equal(t, "test.registry.com/blue:magic", img.RegistryRepoTag())
	assert.Equal(t, "test.registry.com/blue:beefcafe", img.RegistryRepoDigest())
}
