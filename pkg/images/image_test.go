package images

import (
	"os"
	"testing"

	"github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/partition/mbr"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/compression"
	"github.com/stretchr/testify/assert"
)

func setup() {
	compression.SetupWindowsTests()
}

func TestOpensRawImage(t *testing.T) {
	tempDir, err := os.MkdirTemp(os.TempDir(), "raspberrydisk")
	assert.Nil(t, err)
	imageFile := tempDir + "/disk.img"
	createsImageWithOnePartition(t, imageFile)
	tested := Image{filePath: imageFile}

	partitionTable, err := tested.GetPartitionTable()

	assert.Nil(t, err)
	assert.Equal(t, uint32((48*1024*1024)/512), partitionTable.Partitions[0].Size)
}

func TestOpensLz4CompressedImage(t *testing.T) {
	tested := Image{filePath: "../testdata/raspberry_1.0.img.lz4"}

	partitionTable, err := tested.GetPartitionTable()

	assert.Nil(t, err)
	assert.Equal(t, 2, len(partitionTable.Partitions))
}

func TestOpensXzCompressedImage(t *testing.T) {
	tested := Image{filePath: "../testdata/raspberry_1.0.img.xz"}

	partitionTable, err := tested.GetPartitionTable()

	assert.Nil(t, err)
	assert.Equal(t, 2, len(partitionTable.Partitions))
}

// ------------

func createsImageWithOnePartition(t *testing.T, imageFile string) {
	var diskSize int64 = 64 * 1024 * 1024 // 64 MB
	var sectorSize int64 = 512

	mydisk, err := diskfs.Create(imageFile, diskSize, diskfs.Raw, diskfs.SectorSize(sectorSize))
	assert.Nil(t, err)
	startBoot := uint32(1)
	sizeBoot := uint32((48 * 1024 * 1024) / sectorSize) //48mb
	table := &mbr.Table{
		Partitions: []*mbr.Partition{
			{
				Bootable: false,
				Type:     mbr.Type(0x0c), // Fat32LBA
				Start:    startBoot,
				Size:     sizeBoot,
			},
		},
	}

	err = mydisk.Partition(table)
	assert.Nil(t, err)
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}
