package compression

import (
	"bytes"
	"crypto/rand"
	"io"
)

type CompressionTool struct {
	Name                string
	FileExtension       string
	SpeedCompressionArg []string
	HighCompressionArg  []string
}

var db []*CompressionTool = []*CompressionTool{
	{Name: "lz4", FileExtension: "lz4", SpeedCompressionArg: []string{"--fast"}, HighCompressionArg: []string{"--best"}},
	{Name: "gzip", FileExtension: "gz", SpeedCompressionArg: []string{"--fast"}, HighCompressionArg: []string{"--best"}},
	{Name: "xz", FileExtension: "xz", SpeedCompressionArg: []string{"-1"}, HighCompressionArg: []string{"-9"}},
	{Name: "zstd", FileExtension: "zst", SpeedCompressionArg: []string{"--fast", "-T0"}, HighCompressionArg: []string{"-19", "-T0"}},
}

var availableTools map[string]*CompressionTool = nil

func AvailableToolMap() map[string]*CompressionTool {
	if availableTools == nil {
		availableTools = map[string]*CompressionTool{}
		for _, tool := range db {
			if tool.IsAvailable() {
				availableTools[tool.Name] = tool
			}
		}
	}

	return availableTools
}

func AvailableToolByFileExtension() map[string]*CompressionTool {
	result := map[string]*CompressionTool{}
	for _, tool := range AvailableToolMap() {
		result[tool.FileExtension] = tool
	}
	return result
}

func AvailableToolsByFastOrder() []*CompressionTool {
	fastOrder := []string{"lz4", "gzip", "zstd", "xz"}
	return toolsByOrder(fastOrder)
}

func AvailableToolsByHighRatioOrder() []*CompressionTool {
	compressionRatioOrder := []string{"zstd", "zstd", "gzip", "lz4"}
	return toolsByOrder(compressionRatioOrder)
}

func toolsByOrder(order []string) []*CompressionTool {
	availableTools := AvailableToolMap()

	result := []*CompressionTool{}
	for _, name := range order {
		if tool, ok := availableTools[name]; ok {
			result = append(result, tool)
		}
	}

	return result
}

func (c *CompressionTool) IsAvailable() bool {
	var testSize int64 = 1 * 1024 // 1KB
	randBytes, err := io.ReadAll(io.LimitReader(rand.Reader, testSize))
	if err != nil {
		return false
	}

	compressor := NewStreamCompressor(bytes.NewReader(randBytes), c.Name)
	if compressor.Open() != nil {
		return false
	}
	defer compressor.Close()

	decompressor := NewStreamDecompressor(compressor, c.Name)
	if decompressor.Open() != nil {
		return false
	}
	defer decompressor.Close()

	bufferRead, err := io.ReadAll(decompressor)
	if err != nil {
		return false
	}

	return bytes.Equal(bufferRead, randBytes)
}
