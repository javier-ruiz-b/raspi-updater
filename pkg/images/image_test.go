package images

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/partition/mbr"
	"github.com/stretchr/testify/assert"
)

func setup() {
	if runtime.GOOS == "windows" {
		path := os.Getenv("PATH")
		tools_win_dir, _ := filepath.Abs("../../tools_win")
		os.Setenv("PATH", path+";"+tools_win_dir)
	}
}

func TestOpensRawImage(t *testing.T) {
	tempDir, err := os.MkdirTemp(os.TempDir(), "raspberrydisk")
	assert.Nil(t, err)
	imageFile := tempDir + "/disk.img"
	createsImageWithOnePartition(t, imageFile)
	tested := Image{filePath: imageFile}

	disk, err := tested.ReadDisk()

	assert.Nil(t, err)
	assert.Equal(t, uint32((48*1024*1024)/512), disk.GetPartitionTable().Partitions[0].Size)
}

func TestOpensLz4CompressedImage(t *testing.T) {
	tested := Image{filePath: "../testdata/acceptance_1.0.img.lz4"}

	disk, err := tested.ReadDisk()

	assert.Nil(t, err)
	assert.Equal(t, 2, len(disk.GetPartitionTable().Partitions))
}

func TestOpensXzCompressedImage(t *testing.T) {
	tested := Image{filePath: "../testdata/raspberry_1.0.img.xz"}

	disk, err := tested.ReadDisk()

	assert.Nil(t, err)
	assert.Equal(t, 2, len(disk.GetPartitionTable().Partitions))
}

// ------------

func createsImageWithOnePartition(t *testing.T, imageFile string) {
	var diskSize int64 = 64 * 1024 * 1024 // 64 MB
	var sectorSize int64 = 512

	mydisk, err := diskfs.Create(imageFile, diskSize, diskfs.Raw)
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
