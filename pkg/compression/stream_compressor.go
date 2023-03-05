package compression

import (
	"io"
)

func NewStreamCompressor(inStream io.Reader, outStream io.Writer, compressor string) *CompressionStream {
	return NewStreamCompressorN(inStream, -1, outStream, compressor)
}

func NewStreamCompressorN(inStream io.Reader, sizeIn int64, outStream io.Writer, compressor string) *CompressionStream {
	return &CompressionStream{
		inStream:    inStream,
		sizeIn:      sizeIn,
		outStream:   outStream,
		sizeOut:     -1,
		command:     compressor,
		commandArgs: []string{"-c", "-"},
	}
}
