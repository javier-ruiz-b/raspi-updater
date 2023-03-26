package images

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ImageDir struct {
	imagesDir string
	backupDir string
}

func NewImageDir(imagesDir string) *ImageDir {
	backupDir := imagesDir + "/backup"
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		log.Println("Cannot create ", backupDir, " ", err)
	}
	return &ImageDir{
		imagesDir: imagesDir,
		backupDir: backupDir,
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

func (i *ImageDir) CreateBackup(imageName string, version string, compression string) (*os.File, error) {
	fileName := imageName + "_" + version + "_" + time.Now().Format("2006-01-02_15-04-05") + ".img." + compression
	return os.Create(i.backupDir + "/" + fileName)
}

func (i *ImageDir) BackupExists(imageName string, version string) bool {
	matches, err := filepath.Glob(i.backupDir + "/" + imageName + "_" + version + "_*.img*")
	return err == nil && len(matches) > 0
}
