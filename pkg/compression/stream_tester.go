package compression

import "io"

func NewStreamTester(inStream io.Reader, compressor string) *CompressionStream {
	return &CompressionStream{
		inStream:    inStream,
		sizeIn:      -1,
		command:     compressor,
		commandArgs: []string{"-t", "-"},
	}
}
