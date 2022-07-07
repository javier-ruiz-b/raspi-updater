package test

import (
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"testing"

	"github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/partition/mbr"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/client"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/config"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/server"
	"github.com/stretchr/testify/assert"
)

var address string
var tempDir string
var clientImage string
var serv *server.Server

func setup() {
	address = "localhost:25469"
	serv = server.NewServer(newServerConfig())
	go serv.Listen()
	runtime.Gosched()

	var err error
	tempDir, err = ioutil.TempDir(os.TempDir(), "acceptance")
	if err != nil {
		log.Panic(err)
	}

	clientImage := tempDir + "/client.img"
	err = createEmptyImage(clientImage, 64*1024*1024)
	if err != nil {
		log.Panic(err)
	}

}

func teardown() {
	defer os.RemoveAll(tempDir)
	serv.Close()
}

func TestUpdateClientBinary(t *testing.T) {
	options := newClientConfig()
	differentVersion := "0.0.0"
	options.Version = &differentVersion

	err := client.RunClient(options)

	assert.EqualError(t, err, "")
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

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

// helpers

func newClientConfig() *config.ClientConfig {
	result := config.NewClientConfig()

	id := "acceptance"
	result.Id = &id
	result.DiskDevice = &clientImage

	return result
}

func newServerConfig() *config.ServerConfig {
	result := config.NewServerConfig()

	imagesDir := "../testdata/images"
	updaterDir := "../testdata/bin"
	result.ImagesDir = &imagesDir
	result.UpdaterDir = &updaterDir

	return result
}
