package images

import (
	"errors"
	"io"
	"os"
	"strings"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/compression"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/disk"
)

type Image struct {
	filePath string
}

func (i *Image) Name() string {
	return i.filePath
}

func (i *Image) OpenImage() (io.ReadSeekCloser, error) {
	fileInfo, err := os.Stat(i.filePath)
	if err != nil {
		return nil, err
	}

	inStream, err := os.Open(i.filePath)
	if err != nil {
		return inStream, err
	}

	compressionTool := ""
	name := strings.ToLower(fileInfo.Name())
	switch {
	case strings.HasSuffix(name, ".img"):
		return inStream, nil
	case strings.HasSuffix(name, ".img.lz4"):
		compressionTool = "lz4"
	case strings.HasSuffix(name, ".img.xz"):
		compressionTool = "xz"
	case strings.HasSuffix(name, ".img.zstd"):
		compressionTool = "zstd"
	case strings.HasSuffix(name, ".img.gz"):
		compressionTool = "gzip"
	default:
		return nil, errors.New("Unknown file extension: " + fileInfo.Name())
	}

	stream := compression.NewStreamDecompressor(inStream, compressionTool)
	return stream, stream.Open()
}

// func commandOutputToPipe(name string, args ...string) (io.ReadCloser, error) {
// 	cmd := exec.Command(name, args...)
// 	stdout, err := cmd.StdoutPipe()
// 	if err != nil {
// 		return nil, err
// 	}
// 	go cmd.Run()
// 	return stdout, nil
// }

func (i *Image) GetPartitionTable() (*disk.PartitionTable, error) {
	image, err := i.OpenImage()
	if err != nil {
		return nil, err
	}
	defer image.Close()

	tempFile, err := os.CreateTemp(os.TempDir(), "imagembr")
	if err != nil {
		return nil, err
	}

	mbrPartitionTableSize := 512
	_, err = io.CopyN(tempFile, image, int64(mbrPartitionTableSize))
	if err != nil {
		return nil, err
	}
	tempFile.Close()

	result := disk.NewDisk(tempFile.Name())
	err = result.Read()
	if err != nil {
		return nil, err
	}

	return result.GetPartitionTable(), nil
}
