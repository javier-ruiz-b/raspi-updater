package compression

import (
	"io"
)

func NewStreamCompressor(inStream io.Reader, tool *CompressionTool) *CompressionStream {
	return NewStreamCompressorN(inStream, -1, tool)
}

func NewStreamCompressorN(inStream io.Reader, sizeIn int64, tool *CompressionTool) *CompressionStream {
	return &CompressionStream{
		inStream: inStream,
		sizeIn:   sizeIn,
		tool:     tool,
	}
}
