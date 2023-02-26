package disk

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	diskfs "github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/partition/mbr"
)

type Disk struct {
	diskDevice     string
	partitionTable *PartitionTable
}

func NewDisk(diskDevice string) *Disk {
	return &Disk{
		diskDevice: diskDevice,
	}
}

func (d *Disk) Read() error {
	disk, err := diskfs.Open(d.diskDevice)
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

	result := &PartitionTable{
		Size:       uint64(disk.Size),
		SectorSize: int(disk.LogicalBlocksize),
		Partitions: []Partition{},
	}

	for i, part := range mbrTable.GetPartitions() {
		mbrPart, ok := part.(*mbr.Partition)
		if !ok {
			return fmt.Errorf("partition index %d is not a MBR partition", i)
		}
		if mbrPart.Type == 0x0 {
			continue
		}

		result.Partitions = append(result.Partitions, Partition{
			Type:   PartitionType(mbrPart.Type),
			Start:  mbrPart.Start,
			Size:   mbrPart.Size,
			index:  i,
			parent: d,
		})
	}

	d.partitionTable = result

	return nil
}

func (d *Disk) Write() error {
	disk, err := diskfs.Open(d.diskDevice)
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

	if mbrTable.LogicalSectorSize != d.partitionTable.SectorSize {
		return fmt.Errorf("sector size is not equal. Read sector size: %d, Write sector size: %d",
			mbrTable.LogicalSectorSize, d.partitionTable.SectorSize)
	}

	for i, partition := range mbrTable.Partitions {
		if i >= len(d.partitionTable.Partitions) {
			partition.Type = 0
			partition.Start = 0
			partition.Size = 0
			continue
		}

		newPartition := d.partitionTable.Partitions[i]
		partition.Type = mbr.Type(newPartition.Type)
		partition.Start = newPartition.Start
		partition.Size = newPartition.Size
	}

	disk.Partition(mbrTable)

	return nil
}

func (d *Disk) ReadVersion() (string, error) {
	if d.partitionTable == nil {
		return "", fmt.Errorf("invalid partition table")
	}

	var bootPartition *Partition
	for _, partition := range d.partitionTable.Partitions {
		if isFatPartition(&partition) {
			bootPartition = &partition
			break
		}
	}
	if bootPartition == nil {
		return "", fmt.Errorf("boot partition not found")
	}

	versionFile, err := bootPartition.OpenFile("/version", os.O_RDONLY)
	if err != nil {
		return "", err
	}

	out := &bytes.Buffer{}
	_, err = io.Copy(out, versionFile)
	return out.String(), err
}

func (d *Disk) GetPartitionTable() *PartitionTable {
	return d.partitionTable
}

func (d *Disk) MergePartitionTable(desired *PartitionTable) error {
	result, err := mergePartitionTables(desired, d.partitionTable)
	if err == nil {
		d.partitionTable = result
	}
	return err
}
