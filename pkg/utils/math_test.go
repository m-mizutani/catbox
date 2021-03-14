package utils_test

import (
	"testing"

	"github.com/m-mizutani/catbox/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestMin(t *testing.T) {
	assert.Equal(t, 0, utils.Min(0, 1))
	assert.Equal(t, 0, utils.Min(1, 0))
	assert.Equal(t, 2, utils.Min(2, 2))
}
