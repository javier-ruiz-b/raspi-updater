package images

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/disk"
)

type Image struct {
	filePath string
}

func (i *Image) Name() string {
	return i.filePath
}

func (i *Image) OpenImage() (io.ReadCloser, error) {
	fileInfo, err := os.Stat(i.filePath)
	if err != nil {
		return nil, err
	}
	name := strings.ToLower(fileInfo.Name())
	switch {
	case strings.HasSuffix(name, ".img"):
		return os.Open(i.filePath)
	case strings.HasSuffix(name, ".img.lz4"):
		return commandOutputToPipe("lz4", "-dc", i.filePath)
	case strings.HasSuffix(name, ".img.xz"):
		return commandOutputToPipe("xz", "-dc", i.filePath)
	case strings.HasSuffix(name, ".img.zstd"):
		return commandOutputToPipe("zstd", "-dc", i.filePath)
	case strings.HasSuffix(name, ".img.gz"):
		return commandOutputToPipe("gzip", "-dc", i.filePath)
	default:
		return nil, errors.New("Unknown file extension: " + fileInfo.Name())
	}
}

func commandOutputToPipe(name string, args ...string) (io.ReadCloser, error) {
	cmd := exec.Command(name, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	go cmd.Run()
	return stdout, nil
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
