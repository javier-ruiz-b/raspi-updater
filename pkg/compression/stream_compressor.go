package compression

import (
	"io"
)

func NewStreamCompressor(outStream io.Writer, inStream io.Reader, compressor string) *CompressionStream {
	return NewStreamCompressorN(outStream, inStream, -1, compressor)
}

func NewStreamCompressorN(outStream io.Writer, inStream io.Reader, sizeIn int64, compressor string) *CompressionStream {
	return &CompressionStream{
		inStream:    inStream,
		sizeIn:      sizeIn,
		outStream:   outStream,
		sizeOut:     -1,
		command:     compressor,
		commandArgs: []string{"-c", "-"},
	}
}
