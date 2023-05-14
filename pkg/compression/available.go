package compression

import (
	"bytes"
	"crypto/rand"
	"errors"
	"io"
)

type CompressionTool struct {
	Name            string
	Binary          string
	FileExtension   string
	CompressionArgs []string
}

var db []*CompressionTool = []*CompressionTool{
	{Name: "lz4fast", Binary: "lz4", FileExtension: "lz4", CompressionArgs: []string{"--fast"}},
	{Name: "lz4best", Binary: "lz4", FileExtension: "lz4", CompressionArgs: []string{"--best"}},
	{Name: "gzip", Binary: "gzip", FileExtension: "gz", CompressionArgs: []string{"--best"}},
	{Name: "xz", Binary: "xz", FileExtension: "xz", CompressionArgs: []string{"-6"}},
	{Name: "zstdfast", Binary: "zstd", FileExtension: "zst", CompressionArgs: []string{"-1", "-T0"}},
	{Name: "zstdbest", Binary: "zstd", FileExtension: "zst", CompressionArgs: []string{"-19", "-T0"}},
}

var availableTools map[string]*CompressionTool = nil

func MatchAndGetFastestAndBestRatioCompresors(serverCompressors []string) (*CompressionTool, *CompressionTool, error) {
	availableTools := AvailableToolMap()
	matchingCompressors := map[string]*CompressionTool{}
	for _, name := range serverCompressors {
		if tool, ok := availableTools[name]; ok {
			matchingCompressors[name] = tool
		}
	}

	if len(matchingCompressors) == 0 {
		return nil, nil, errors.New("no matching compressors found")
	}

	fastToRatioOrder := []string{"lz4fast", "lz4best", "zstdfast", "gzip", "xz", "zstdbest"}

	fastTool := getFirstMatchingCompressor(matchingCompressors, fastToRatioOrder)
	highRatioTool := getFirstMatchingCompressor(matchingCompressors, reverseArray(fastToRatioOrder))

	return fastTool, highRatioTool, nil
}

func reverseArray(arr []string) []string {
	result := arr
	for i := 0; i < len(result)/2; i++ {
		// Swap elements at positions i and len(result)-1-i
		result[i], result[len(result)-1-i] = result[len(result)-1-i], result[i]
	}
	return result
}

func getFirstMatchingCompressor(compressors map[string]*CompressionTool, matchingOrder []string) *CompressionTool {
	for _, name := range matchingOrder {
		if tool, ok := compressors[name]; ok {
			return tool
		}
	}

	//no matches. Fallback: take any
	for _, tool := range compressors {
		return tool
	}

	return nil
}

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

func (c *CompressionTool) IsAvailable() bool {
	var testSize int64 = 1 * 1024 // 1KB
	randBytes, err := io.ReadAll(io.LimitReader(rand.Reader, testSize))
	if err != nil {
		return false
	}

	compressor := NewStreamCompressor(bytes.NewReader(randBytes), c.Binary)
	if compressor.Open() != nil {
		return false
	}
	defer compressor.Close()

	decompressor := NewStreamDecompressor(compressor, c.Binary)
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
