package compression

import "io"

func NewStreamDecompressor(inStream io.ReadCloser, compressor string) *CompressionStream {
	return &CompressionStream{
		inStream:    inStream,
		sizeIn:      -1,
		command:     compressor,
		commandArgs: []string{"-dc", "-"},
	}
}
