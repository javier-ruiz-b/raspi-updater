package compression

import "io"

func NewStreamTester(inStream io.Reader, tool *CompressionTool) *CompressionStream {
	return &CompressionStream{
		inStream:  inStream,
		sizeIn:    -1,
		tool:      tool,
		extraArgs: []string{"-t", "-"},
	}
}
