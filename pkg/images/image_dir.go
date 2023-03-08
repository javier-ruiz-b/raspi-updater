package images

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ImageDir struct {
	imagesDir string
}

func NewImageDir(imagesDir string) *ImageDir {
	return &ImageDir{
		imagesDir: imagesDir,
	}
}

func (i *ImageDir) FindImage(imageName string) (*Image, error) {
	matches, err := filepath.Glob(i.imagesDir + "/" + imageName + "_*.img*")
	if err != nil {
		return nil, err
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no matching image found for %s", imageName)
	}

	if len(matches) != 1 {
		return nil, fmt.Errorf("there are %d images matching %s", len(matches), strings.Join(matches, " "))
	}

	_, err = os.Stat(matches[0])
	if err != nil {
		return nil, err
	}

	return &Image{
		filePath: matches[0],
	}, nil
}
