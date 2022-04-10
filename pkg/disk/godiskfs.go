package disk

import (
	"errors"
	"fmt"

	diskfs "github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/partition/mbr"
)

func ReadDisk(diskDevice string) (*PartitionTable, error) {
	disk, err := diskfs.Open(diskDevice)
	if err != nil {
		return nil, err
	}

	table, err := disk.GetPartitionTable()
	if err != nil {
		return nil, err
	}

	mbrTable, ok := table.(*mbr.Table)
	if !ok {
		return nil, errors.New("not a MBR partition table")
	}

	result := &PartitionTable{
		Size:       uint64(disk.Size),
		SectorSize: int(disk.LogicalBlocksize),
		Partitions: []Partition{},
	}

	for i, part := range mbrTable.GetPartitions() {
		mbrPart, ok := part.(*mbr.Partition)
		if !ok {
			return nil, fmt.Errorf("partition index %d is not a MBR partition", i)
		}
		if mbrPart.Type == 0x0 {
			continue
		}

		result.Partitions = append(result.Partitions, Partition{
			Type:  PartitionType(mbrPart.Type),
			Start: mbrPart.Start,
			Size:  mbrPart.Size,
		})
	}
	return result, nil
}

func EditDisk(diskDevice string, newTable *PartitionTable) error {
	disk, err := diskfs.Open(diskDevice)
	if err != nil {
		return err
	}

	table, err := disk.GetPartitionTable()
	if err != nil {
		return err
	}

	mbrTable, ok := table.(*mbr.Table)
	if !ok {
		return errors.New("not a MBR partition table")
	}

	if mbrTable.LogicalSectorSize != newTable.SectorSize {
		return fmt.Errorf("sector size is not equal. Read sector size: %d, Write sector size: %d",
			mbrTable.LogicalSectorSize, newTable.SectorSize)
	}

	for i, partition := range mbrTable.Partitions {
		if i >= len(newTable.Partitions) {
			partition.Type = 0
			partition.Start = 0
			partition.Size = 0
			continue
		}

		newPartition := newTable.Partitions[i]
		partition.Type = mbr.Type(newPartition.Type)
		partition.Start = newPartition.Start
		partition.Size = newPartition.Size
	}

	disk.Partition(mbrTable)

	return nil
}

// func check(err error) {
// 	log.Panic(err)
// }

// func Create(diskDevice string) {
// 	if diskDevice == "" {
// 		log.Fatal("must have a valid path for diskImg")
// 	}
// 	var diskSize int64 = 10 * 1024 * 1024 // 10 MB
// 	//sectorSize := 512
// 	mydisk, err := diskfs.Create(diskDevice, diskSize, diskfs.Raw)
// 	check(err)

// 	// the following line is required for an ISO, which may have logical block sizes
// 	// only of 2048, 4096, 8192
// 	mydisk.LogicalBlocksize = 2048
// 	fspec := disk.FilesystemSpec{Partition: 0, FSType: filesystem.TypeISO9660, VolumeLabel: "label"}
// 	fs, err := mydisk.CreateFilesystem(fspec)
// 	check(err)
// 	rw, err := fs.OpenFile("demo.txt", os.O_CREATE|os.O_RDWR)
// 	content := []byte("demo")
// 	_, err = rw.Write(content)
// 	check(err)
// 	iso, ok := fs.(*iso9660.FileSystem)
// 	if !ok {
// 		// check(fmt.Errorf("not an iso9660 filesystem"))
// 	}
// 	err = iso.Finalize(iso9660.FinalizeOptions{})
// 	check(err)
// }

// // // create a disk image
// // diskImg := "/tmp/disk.img"
// // disk := diskfs.Create(diskImg, diskSize, diskfs.Raw, diskfs.SectorSizeDefault)
