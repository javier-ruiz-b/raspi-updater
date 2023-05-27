package compression

import "io"

func NewStreamDecompressor(inStream io.Reader, tool *CompressionTool) *CompressionStream {
	return &CompressionStream{
		inStream:  inStream,
		sizeIn:    -1,
		tool:      tool,
		extraArgs: []string{"-dc", "-"},
	}
}
