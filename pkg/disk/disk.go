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
	defer disk.File.Close()

	table, err := disk.GetPartitionTable()
	if err != nil {
		return err
	}

	mbrTable, ok := table.(*mbr.Table)
	if !ok {
		return errors.New("not a MBR partition table")
	}

	partitions := []Partition{}
	for i, part := range mbrTable.GetPartitions() {
		mbrPart, ok := part.(*mbr.Partition)
		if !ok {
			return fmt.Errorf("partition index %d is not a MBR partition", i)
		}
		if mbrPart.Type == 0x0 {
			continue
		}

		partitions = append(partitions, Partition{
			Type:   PartitionType(mbrPart.Type),
			Start:  mbrPart.Start,
			Size:   mbrPart.Size,
			Index:  i,
			parent: d,
		})
	}
	size := uint64(disk.Size)
	if len(partitions) > 0 {
		lastSectorBytes := partitions[len(partitions)-1].EndSector() * uint64(disk.LogicalBlocksize)
		size = Max(size, lastSectorBytes)
	}
	d.partitionTable = &PartitionTable{
		Size:       size,
		SectorSize: int(disk.LogicalBlocksize),
		Partitions: partitions,
	}
	return nil
}

func (d *Disk) Write() error {
	disk, err := diskfs.Open(d.diskDevice)
	if err != nil {
		return err
	}
	defer disk.File.Close()

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

	return disk.Partition(mbrTable)
}

func (d *Disk) ReadVersion() (string, error) {
	bootPartition, err := d.GetBootPartition()
	if err != nil {
		return "", err
	}

	versionFile, err := bootPartition.OpenFile("/version", os.O_RDONLY)
	if err != nil {
		return "", err
	}
	defer versionFile.Close()

	out := &bytes.Buffer{}
	_, err = io.Copy(out, versionFile)

	return out.String(), err
}

func (d *Disk) WriteVersion(version string) error {
	bootPartition, err := d.GetBootPartition()
	if err != nil {
		return err
	}

	versionFile, err := bootPartition.OpenFile("/version", os.O_CREATE|os.O_RDWR)
	if err != nil {
		return err
	}
	defer versionFile.Close()

	_, err = io.Copy(versionFile, bytes.NewBuffer([]byte(version)))

	return err
}

func (d *Disk) GetBootPartition() (*Partition, error) {
	if d.partitionTable == nil {
		return nil, fmt.Errorf("invalid partition table")
	}

	var bootPartition *Partition
	for _, partition := range d.partitionTable.Partitions {
		if isFatPartition(&partition) {
			bootPartition = &partition
			break
		}
	}
	if bootPartition == nil {
		return nil, fmt.Errorf("boot partition not found")
	}
	return bootPartition, nil
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

func Max(x, y uint64) uint64 {
	if x < y {
		return y
	}
	return x
}
