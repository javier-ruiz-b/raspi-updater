package compression

import "io"

func NewStreamDecompressor(outStream io.Writer, inStream io.Reader, compressor string) *CompressionStream {
	return NewStreamDecompressorN(outStream, -1, inStream, compressor)
}

func NewStreamDecompressorN(outStream io.Writer, sizeOut int64, inStream io.Reader, compressor string) *CompressionStream {
	return &CompressionStream{
		inStream:    inStream,
		sizeIn:      -1,
		outStream:   outStream,
		sizeOut:     sizeOut,
		command:     compressor,
		commandArgs: []string{"-dc", "-"},
	}
}
