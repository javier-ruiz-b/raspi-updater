package compression

import "io"

func NewStreamDecompressor(inStream io.Reader, outStream io.Writer, compressor string) *CompressionStream {
	return NewStreamDecompressorN(inStream, outStream, -1, compressor)
}

func NewStreamDecompressorN(inStream io.Reader, outStream io.Writer, sizeOut int64, compressor string) *CompressionStream {
	return &CompressionStream{
		inStream:    inStream,
		sizeIn:      -1,
		outStream:   outStream,
		sizeOut:     sizeOut,
		command:     compressor,
		commandArgs: []string{"-dc", "-"},
	}
}
