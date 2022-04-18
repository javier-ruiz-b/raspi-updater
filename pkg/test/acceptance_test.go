package test

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"testing"

	"github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/partition/mbr"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/client"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/server"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/version"
	"github.com/stretchr/testify/assert"
)

var address string
var tempDir string
var clientImage string
var serv *server.Server

func setup() {
	address = "localhost:25469"
	serv = server.NewServer(address, "test/images")
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

func TestUpdater(t *testing.T) {
	err := client.RunClient(newConfigWithDifferentVersion())
	assert.Equal(t, err, io.EOF)
}

func newConfigWithDifferentVersion() *client.Config {
	result := config()
	result.Version = "0.0.0"
	return result
}

func config() *client.Config {
	return &client.Config{
		ServerAddress: address,
		Id:            "acceptance",
		DiskDevice:    clientImage,
		Version:       version.VERSION,
	}
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
