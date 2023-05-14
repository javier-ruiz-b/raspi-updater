package images

import (
	"errors"
	"io"
	"os"
	"regexp"
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

	name := strings.ToLower(fileInfo.Name())
	if strings.HasSuffix(name, ".img") {
		return inStream, nil
	}

	extensionRe := regexp.MustCompile(`.*\.img\.(.*)`)
	extensionMatch := extensionRe.FindStringSubmatch(name)
	if len(extensionMatch) != 2 {
		return nil, errors.New("Unknown file extension: " + fileInfo.Name())
	}

	tool, found := compression.AvailableToolByFileExtension()[extensionMatch[1]]
	if !found {
		return nil, errors.New("No compression tool found for file extension " + extensionMatch[1])
	}

	stream := compression.NewStreamDecompressor(inStream, tool.Binary)
	return stream, stream.Open()
}

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
