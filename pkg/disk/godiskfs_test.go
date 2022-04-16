package disk

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	diskfs "github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/partition/mbr"
	"github.com/stretchr/testify/assert"
)

var tempDir string
var imageFile string

func setup() {
	tempDir, err := ioutil.TempDir(os.TempDir(), "raspberrydisk")
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
	table, err := ReadDisk(imageFile)

	assert.Nil(t, err)
	assert.Equal(t, &PartitionTable{
		Partitions: []Partition{
			{
				Type:  Fat32LBA,
				Start: 1,
				Size:  (48 * 1024 * 1024) / 512,
			},
			{
				Type:  Linux,
				Start: 1 + (48*1024*1024)/512,
				Size:  (8 * 1024 * 1024) / 512,
			},
		},
		SectorSize: 512,
		Size:       64 * 1024 * 1024,
	}, table)
}

func TestMergesWithDesiredTable(t *testing.T) {
	table, err := ReadDisk(imageFile)
	assert.Nil(t, err)

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

	resultTable, err := MergeTables(desiredTable, table)
	assert.Nil(t, err)

	err = EditDisk(imageFile, resultTable)
	assert.Nil(t, err)

	table, err = ReadDisk(imageFile)
	assert.Nil(t, err)

	originalStartBoot := uint32(1)
	originalSizeBoot := uint32((48 * 1024 * 1024) / sectorSize) //48mb
	assert.Equal(t, &PartitionTable{
		Partitions: []Partition{
			{
				Type:  Fat32LBA,
				Start: originalStartBoot,
				Size:  originalSizeBoot,
			},
			{
				Type:  Linux,
				Start: uint32(originalStartBoot + originalSizeBoot),
				Size:  uint32(rootSize),
			},
			{
				Type:  Linux,
				Start: uint32(originalStartBoot + originalSizeBoot + uint32(rootSize)),
				Size:  uint32(dataSize),
			},
		},
		Size:       64 * 1024 * 1024,
		SectorSize: sectorSize,
	}, table)
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
