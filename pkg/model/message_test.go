package model_test

import (
	"testing"

	"github.com/m-mizutani/catbox/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestQueue(t *testing.T) {
	msg := model.ScanRequestMessage{
		OutS3Prefix: "test/path/to",
	}
	// Do not append slash between prefix and key automatically (it's confusing)
	assert.Equal(t, "test/path/totrivy.json.gz", msg.S3Key(model.ScannerTrivy))
}
