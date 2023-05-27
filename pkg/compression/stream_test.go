package compression

import (
	"bytes"
	"crypto/rand"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setup() {
	SetupWindowsTests()
}

func TestStreamsLimited1MbOfRandomData(t *testing.T) {
	randBytes, err := io.ReadAll(io.LimitReader(rand.Reader, 1024*1024))
	assert.Nil(t, err)
	assert.Equal(t, 1024*1024, len(randBytes))
	tool := AvailableToolMap()["lz4fast"]

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

	tool := AvailableToolMap()["lz4fast"]
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
