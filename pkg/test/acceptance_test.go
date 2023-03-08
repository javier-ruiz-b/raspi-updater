package test

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/partition/mbr"
	"github.com/hlubek/readercomp"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/client"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/config"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/images"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/runner"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/server"
	"github.com/stretchr/testify/assert"
)

var address string
var tempDir string
var clientImage string
var serv *server.Server
var clientConfig *config.ClientConfig
var imagesDir string = "../testdata"
var id string = "acceptance"

func setup() {
	if runtime.GOOS == "windows" {
		path := os.Getenv("PATH")
		tools_win_dir, _ := filepath.Abs("../../tools_win")
		os.Setenv("PATH", path+";"+tools_win_dir)
	}

	address = "localhost:25469"
	serverConfig := newServerConfig()
	serverConfig.ImagesDir = &imagesDir
	serv = server.NewServer(serverConfig)
	go serv.Listen()
	runtime.Gosched()

	var err error
	tempDir, err = os.MkdirTemp(os.TempDir(), "acceptance")
	if err != nil {
		log.Panic(err)
	}

	clientImage = tempDir + "/client.img"
	err = createEmptyImage(clientImage, 64*1024*1024)
	if err != nil {
		log.Panic(err)
	}

	clientConfig = newClientConfig()
	clientConfig.DiskDevice = &clientImage
}

func teardown() {
	serv.Close()
	os.RemoveAll(tempDir)
}

func TestUpdateClientBinary(t *testing.T) {
	runner := clientConfig.Runner.(*runner.FakeRunner)
	differentVersion := "0.0.0"
	clientConfig.Version = &differentVersion

	err := client.RunClient(clientConfig)

	assert.True(t, runner.IsRun())
	assert.Nil(t, err)
}

func TestAcceptance(t *testing.T) {
	runner := clientConfig.Runner.(*runner.FakeRunner)

	err := client.RunClient(clientConfig)

	assert.True(t, runner.IsRun())
	assert.Nil(t, err)
	image, err := images.NewImageDir(imagesDir).FindImage(id)
	assert.Nil(t, err)
	imageStream, err := image.OpenImage()
	assert.Nil(t, err)
	diskStream, err := os.Open(clientImage)
	assert.Nil(t, err)
	_, err = imageStream.Seek(512*2048, 1) // skip to first partition
	assert.Nil(t, err)
	_, err = diskStream.Seek(512*2048, 1)
	assert.Nil(t, err)
	result, err := readercomp.Equal(imageStream, diskStream, 1024*1024)
	assert.Nil(t, err)
	log.Print(clientImage)
	assert.True(t, result, "Disk contents are not equal")
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

	compressionTool := "lz4"
	result.Id = &id
	result.DiskDevice = &clientImage
	result.Runner = runner.NewFakeRunner()
	result.CompressionTool = &compressionTool

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
