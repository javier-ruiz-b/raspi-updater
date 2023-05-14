package compression

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompressionToolLz4IsAvailable(t *testing.T) {
	tested := &CompressionTool{
		Name: "../../tools_win/lz4",
	}

	assert.True(t, tested.IsAvailable())
}

func TestInexistentCompressionToolIsNotAvailable(t *testing.T) {
	tested := &CompressionTool{
		Name: "../../tools_win/wrong.exe",
	}

	assert.False(t, tested.IsAvailable())
}
