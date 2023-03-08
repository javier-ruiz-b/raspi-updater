package compression

import (
	"io"
)

func NewStreamCompressor(inStream io.Reader, compressor string) *CompressionStream {
	return NewStreamCompressorN(inStream, -1, compressor)
}

func NewStreamCompressorN(inStream io.Reader, sizeIn int64, compressor string) *CompressionStream {
	return &CompressionStream{
		inStream:    inStream,
		sizeIn:      sizeIn,
		command:     compressor,
		commandArgs: []string{"-c", "-"},
	}
}
