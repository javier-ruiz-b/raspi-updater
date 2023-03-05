package disk

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"testing"

	diskfs "github.com/diskfs/go-diskfs"
	diskfsdisk "github.com/diskfs/go-diskfs/disk"
	"github.com/diskfs/go-diskfs/filesystem"
	"github.com/diskfs/go-diskfs/partition/mbr"
	"github.com/stretchr/testify/assert"
)

var tempDir string
var imageFile string

func setup() {
	tempDir, err := os.MkdirTemp(os.TempDir(), "raspberrydisk")
	check(err)

	imageFile = tempDir + "/disk.img"

	var diskSize int64 = 64 * 1024 * 1024 // 64 MB
	mydisk, err := diskfs.Create(imageFile, diskSize, diskfs.Raw)
	check(err)

	var sectorSize int64 = 512
	startBoot := uint32(1)
	sizeBoot := uint32((48 * 1024 * 1024) / sectorSize) //48mb
	startRoot := startBoot + sizeBoot
	sizeRoot := uint32((8 * 1024 * 1024) / sectorSize) //8mb

	mydisk.LogicalBlocksize = sectorSize
	table := &mbr.Table{
		Partitions: []*mbr.Partition{
			{
				Bootable: false,
				Type:     mbr.Type(Fat32LBA),
				Start:    startBoot,
				Size:     sizeBoot,
			},
			{
				Bootable: false,
				Type:     mbr.Type(Linux),
				Start:    startRoot,
				Size:     sizeRoot,
			},
		},
	}

	err = mydisk.Partition(table)
	check(err)
}

func teardown() {
	os.RemoveAll(tempDir)
}

func TestReadsPartitionTable(t *testing.T) {
	tested := NewDisk(imageFile)

	err := tested.Read()

	assert.Nil(t, err)
	assert.Equal(t, 512, tested.partitionTable.SectorSize)
	assert.Equal(t, 64*1024*1024, int(tested.partitionTable.Size))
	assert.Equal(t, 2, len(tested.partitionTable.Partitions))

	assert.Equal(t, Fat32LBA, tested.partitionTable.Partitions[0].Type)
	assert.Equal(t, 1, int(tested.partitionTable.Partitions[0].Start))
	assert.Equal(t, (48*1024*1024)/512, int(tested.partitionTable.Partitions[0].Size))

	assert.Equal(t, Linux, tested.partitionTable.Partitions[1].Type)
	assert.Equal(t, 1+(48*1024*1024)/512, int(tested.partitionTable.Partitions[1].Start))
	assert.Equal(t, (8*1024*1024)/512, int(tested.partitionTable.Partitions[1].Size))
}

func TestMergesWithDesiredTable(t *testing.T) {
	sectorSize := 512

	bootStart := 32
	bootSize := (36 * 1024 * 1024) / sectorSize

	rootStart := bootStart + bootSize
	rootSize := (4 * 1024 * 1024) / sectorSize

	dataStart := rootStart + rootSize
	dataSize := (2 * 1024 * 1024) / sectorSize

	desiredTable := &PartitionTable{
		Partitions: []Partition{
			{
				Type:  Fat32LBA,
				Start: uint32(bootStart),
				Size:  uint32(bootSize),
			},
			{
				Type:  Linux,
				Start: uint32(rootStart),
				Size:  uint32(rootSize),
			},
			{
				Type:  Linux,
				Start: uint32(dataStart),
				Size:  uint32(dataSize),
			},
		},
		Size:       64 * 1024 * 1024,
		SectorSize: sectorSize,
	}

	tested := NewDisk(imageFile)
	assert.Nil(t, tested.Read())
	assert.Nil(t, tested.MergePartitionTable(desiredTable))
	assert.Nil(t, tested.Write())
	assert.Nil(t, tested.Read())
	fmt.Print(tested.GetPartitionTable().GetInfo())

	originalStartBoot := uint32(1)
	originalSizeBoot := uint32((48 * 1024 * 1024) / sectorSize) //48mb

	assert.Equal(t, 512, tested.partitionTable.SectorSize)
	assert.Equal(t, 64*1024*1024, int(tested.partitionTable.Size))
	assert.Equal(t, 3, len(tested.partitionTable.Partitions))

	assert.Equal(t, Fat32LBA, tested.partitionTable.Partitions[0].Type)
	assert.Equal(t, originalStartBoot, tested.partitionTable.Partitions[0].Start)
	assert.Equal(t, originalSizeBoot, tested.partitionTable.Partitions[0].Size)

	assert.Equal(t, Linux, tested.partitionTable.Partitions[1].Type)
	assert.Equal(t, originalStartBoot+originalSizeBoot, tested.partitionTable.Partitions[1].Start)
	assert.Equal(t, rootSize, int(tested.partitionTable.Partitions[1].Size))

	assert.Equal(t, Linux, tested.partitionTable.Partitions[2].Type)
	assert.Equal(t, originalStartBoot+originalSizeBoot+uint32(rootSize), tested.partitionTable.Partitions[2].Start)
	assert.Equal(t, dataSize, int(tested.partitionTable.Partitions[2].Size))
}
func TestReadsVersion(t *testing.T) {
	disk, err := diskfs.Open(imageFile)
	assert.Nil(t, err)
	fspec := diskfsdisk.FilesystemSpec{Partition: 1, FSType: filesystem.TypeFat32, VolumeLabel: "boot"}
	fs, err := disk.CreateFilesystem(fspec)
	assert.Nil(t, err)
	file, err := fs.OpenFile("/version", os.O_CREATE|os.O_RDWR)
	assert.Nil(t, err)
	_, err = io.Copy(file, bytes.NewBuffer([]byte("1.2.3.4")))
	assert.Nil(t, err)
	file.Close()
	tested := NewDisk(imageFile)
	tested.Read()

	version, err := tested.ReadVersion()

	assert.Nil(t, err)
	assert.Equal(t, "1.2.3.4", version)
}

func TestWritesToAndReadsFromPartition(t *testing.T) {
	tested := NewDisk(imageFile)
	err := tested.Read()
	assert.Nil(t, err)
	stream, err := tested.GetPartitionTable().Partitions[0].OpenStream()
	assert.Nil(t, err)

	stream.Write([]byte("Example"))
	stream.Close()

	imageStream, err := os.OpenFile(imageFile, os.O_RDONLY, 0)
	assert.Nil(t, err)
	_, err = imageStream.Seek(512*1, 0)
	assert.Nil(t, err)

	var buffer bytes.Buffer
	io.CopyN(&buffer, imageStream, 7)
	assert.Equal(t, "Example", buffer.String())
}

func check(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}
