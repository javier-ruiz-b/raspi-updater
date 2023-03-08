package compression

import (
	"io"
)

func NewStreamCompressor(inStream io.ReadCloser, compressor string) *CompressionStream {
	return NewStreamCompressorN(inStream, -1, compressor)
}

func NewStreamCompressorN(inStream io.ReadCloser, sizeIn int64, compressor string) *CompressionStream {
	return &CompressionStream{
		inStream:    inStream,
		sizeIn:      sizeIn,
		command:     compressor,
		commandArgs: []string{"-c", "-"},
	}
}
