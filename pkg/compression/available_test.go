package compression

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompressionToolLz4IsAvailable(t *testing.T) {
	var binary string
	if runtime.GOOS == "windows" {
		binary = "../../tools_win/lz4"
	} else {
		binary = "lz4"
	}
	tested := &CompressionTool{
		Binary: binary,
	}

	assert.True(t, tested.IsAvailable())
}

func TestInexistentCompressionToolIsNotAvailable(t *testing.T) {
	tested := &CompressionTool{
		Name: "../../tools_win/wrong.exe",
	}

	assert.False(t, tested.IsAvailable())
}
