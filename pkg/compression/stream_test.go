package compression

import (
	"bytes"
	"crypto/rand"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setup() {
	if runtime.GOOS == "windows" {
		path := os.Getenv("PATH")
		tools_win_dir, _ := filepath.Abs("../../tools_win")
		os.Setenv("PATH", path+";"+tools_win_dir)
	}
}

func TestStreamsLimited1MbOfRandomData(t *testing.T) {
	randBytes, err := io.ReadAll(io.LimitReader(rand.Reader, 1024*1024))
	assert.Nil(t, err)
	assert.Equal(t, 1024*1024, len(randBytes))

	tool := "lz4"
	testedCompressor := NewStreamCompressorN(bytes.NewReader(randBytes), 1024*1024, tool)
	assert.Nil(t, testedCompressor.Open())
	tested := NewStreamDecompressor(testedCompressor, tool)
	assert.Nil(t, tested.Open())

	bufferRead, err := io.ReadAll(tested)
	assert.Nil(t, err)

	assert.Equal(t, bufferRead, randBytes)
}

func TestStreams1MbOfRandomData(t *testing.T) {
	randBytes, err := io.ReadAll(io.LimitReader(rand.Reader, 1024*1024))
	assert.Nil(t, err)
	assert.Equal(t, 1024*1024, len(randBytes))

	tool := "xz"
	testedCompressor := NewStreamCompressor(bytes.NewReader(randBytes), tool)
	assert.Nil(t, testedCompressor.Open())
	tested := NewStreamDecompressor(testedCompressor, tool)
	assert.Nil(t, tested.Open())

	bufferRead, err := io.ReadAll(tested)
	assert.Nil(t, err)

	assert.Equal(t, bufferRead, randBytes)
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}
