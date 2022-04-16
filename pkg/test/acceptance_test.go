package test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/partition/mbr"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/client"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/server"
	"github.com/stretchr/testify/assert"
)

func TestAcceptance(t *testing.T) {
	address := "localhost:25469"
	go server.Server(address, "test/images")

	tempDir, err := ioutil.TempDir(os.TempDir(), "acceptance")
	assert.Nil(t, err)
	defer os.RemoveAll(tempDir)

	clientImage := tempDir + "/client.img"
	assert.Nil(t, createEmptyImage(clientImage, 64*1024*1024))

	assert.Nil(t, client.Client(&client.Config{
		ServerAddress: address,
		Id:            "acceptance",
		DiskDevice:    clientImage,
	}))
}

func createEmptyImage(imageFile string, size int64) error {
	mydisk, err := diskfs.Create(imageFile, size, diskfs.Raw)
	if err != nil {
		return err
	}

	mydisk.LogicalBlocksize = 512
	table := &mbr.Table{
		Partitions: []*mbr.Partition{},
	}
	return mydisk.Partition(table)
}
