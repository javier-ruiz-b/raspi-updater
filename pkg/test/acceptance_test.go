package test

import (
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/partition/mbr"
	"github.com/hlubek/readercomp"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/client"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/compression"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/config"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/disk"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/runner"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/server"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/testdata"
	"github.com/stretchr/testify/assert"
)

var address string
var tempDir string
var clientImage string
var serverImage string
var serv *server.Server

var id string = "acceptance"

func setup() {
	compression.SetupWindowsTests()

	var err error
	tempDir, err = os.MkdirTemp(os.TempDir(), "acceptance")
	if err != nil {
		log.Panic(err)
	}

	address = "localhost:25469"
	serverConfig := newServerConfig()
	serverConfig.ImagesDir = &tempDir
	serv = server.NewServer(serverConfig)
	go serv.Listen()
	runtime.Gosched()

	clientImage = tempDir + "/client.img"
	serverImage = tempDir + "/" + id + "_1.0.img"

}

func teardown() {
	serv.Close()
	os.RemoveAll(tempDir)
}

func TestUpdateClientBinary(t *testing.T) {
	clientConfig := newClientConfig()
	clientConfig.DiskDevice = &clientImage
	runner := clientConfig.Runner.(*runner.FakeRunner)
	differentVersion := "0.0.0"
	clientConfig.Version = &differentVersion

	err := client.RunClient(clientConfig)

	assert.True(t, runner.IsRun())
	assert.Nil(t, err)
}

func TestAcceptance(t *testing.T) {
	assert.Nil(t, createEmptyImage(clientImage, 64*1024*1024))
	assert.Nil(t, createImageToBeCopied(serverImage))

	clientConfig := newClientConfig()
	clientConfig.DiskDevice = &clientImage
	runner := clientConfig.Runner.(*runner.FakeRunner)

	err := client.RunClient(clientConfig)

	assert.True(t, runner.IsRun())
	assert.Nil(t, err)
	serverStream, err := os.Open(serverImage)
	assert.Nil(t, err)
	clientStream, err := os.Open(clientImage)
	assert.Nil(t, err)
	_, err = serverStream.Seek(512*2048, 1) // skip to first partition
	assert.Nil(t, err)
	_, err = clientStream.Seek(512*2048, 1)
	assert.Nil(t, err)
	bufferSize := 1024 * 1024
	result, err := readercomp.Equal(serverStream, clientStream, bufferSize)
	assert.Nil(t, err)
	log.Print("Images: ", serverImage, " ", clientImage)
	assert.True(t, result, "Disk contents are not equal")

	//check backup
	matches, err := filepath.Glob(tempDir + "/backup/*.img*")
	assert.Nil(t, err)
	assert.Equal(t, len(matches), 1)
	expectedImage := tempDir + "/expected.img"
	createEmptyImage(expectedImage, 64*1024*1024)
	expected, err := os.Open(expectedImage)
	assert.Nil(t, err)
	defer expected.Close()

	compressedBackup, err := os.Open(matches[0])
	assert.Nil(t, err)
	decompressedBackup := compression.NewStreamDecompressor(compressedBackup, "lz4")
	err = decompressedBackup.Open()
	assert.Nil(t, err)

	equal, err := readercomp.Equal(decompressedBackup, expected, bufferSize)
	assert.Nil(t, err)
	assert.True(t, equal)
}

func createEmptyImage(imageFile string, size int64) error {
	mydisk, err := diskfs.Create(imageFile, size, diskfs.Raw, 512)
	if err != nil {
		return err
	}

	mydisk.LogicalBlocksize = 512
	table := &mbr.Table{
		Partitions: []*mbr.Partition{},
	}
	return mydisk.Partition(table)
}

func createImageToBeCopied(imageFile string) error {
	if err := createEmptyImage(imageFile, 64*1024*1024); err != nil {
		return err
	}
	blockSize := 512
	startSector := 2048
	fatPartitionSize := 36 * 1024 * 1024
	linuxPartitionSize := 24 * 1024 * 1024

	diskImage := disk.NewDisk(imageFile)
	if err := diskImage.Read(); err != nil {
		return err
	}
	if err := diskImage.MergePartitionTable(&disk.PartitionTable{
		SectorSize: blockSize,
		Size:       uint64(startSector*blockSize + fatPartitionSize + linuxPartitionSize),
		Partitions: []disk.Partition{
			{
				Type:  disk.Fat32CHS,
				Start: uint32(startSector),
				Size:  uint32(fatPartitionSize / blockSize),
			},
			{
				Type:  disk.Linux,
				Start: uint32(startSector + (fatPartitionSize / blockSize)),
				Size:  uint32(linuxPartitionSize / blockSize),
			},
		},
	}); err != nil {
		return err
	}

	if err := diskImage.Write(); err != nil {
		return err
	}

	if err := diskImage.Read(); err != nil {
		return err
	}

	for i, partition := range diskImage.GetPartitionTable().Partitions {
		stream, err := partition.OpenStream()
		if err != nil {
			return err
		}
		defer stream.Close()

		var b byte = '0' + byte(i+1)
		buf := bytes.NewBuffer([]byte{b})
		if _, err := io.CopyBuffer(stream, io.LimitReader(buf, int64(partition.SizeBytes())), make([]byte, 1024)); err != nil {
			return err
		}
	}
	return nil
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

	result.Id = &id
	result.DiskDevice = &clientImage
	result.Runner = runner.NewFakeRunner()

	cert, key := testdata.GetCertificatePaths()
	result.CertificatePath = &cert
	result.KeyPath = &key

	return result
}

func newServerConfig() *config.ServerConfig {
	result := config.NewServerConfig()

	imagesDir := "../testdata/images"
	updaterDir := "../testdata/bin"
	result.ImagesDir = &imagesDir
	result.UpdaterDir = &updaterDir

	cert, key := testdata.GetCertificatePaths()
	result.CertificatePath = &cert
	result.KeyPath = &key

	return result
}
