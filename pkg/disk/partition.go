package disk

import (
	"fmt"
	"io"
	"io/fs"
	"os"

	diskfs "github.com/diskfs/go-diskfs"
)

type Partition struct {
	Type   PartitionType
	Start  uint32
	Size   uint32
	Index  int
	parent *Disk
}

// Type constants for the GUID for type of partition, see https://en.wikipedia.org/wiki/GUID_Partition_Table#Partition_entries
type PartitionType byte

// List of GUID partition types
const (
	Empty         PartitionType = 0x00
	Fat12         PartitionType = 0x01
	Fat16         PartitionType = 0x04
	Fat16b        PartitionType = 0x06
	Fat32CHS      PartitionType = 0x0b
	Fat32LBA      PartitionType = 0x0c
	Fat16bLBA     PartitionType = 0x0e
	Linux         PartitionType = 0x83
	LinuxExtended PartitionType = 0x85
	LinuxLVM      PartitionType = 0x8e
)

func (p *Partition) EndSector() uint64 {
	return uint64(p.Start + p.Size)
}

func (p *Partition) SizeBytes() uint64 {
	return uint64(p.Size) * uint64(p.parent.partitionTable.SectorSize)
}

func (p *Partition) ReadDir(path string) ([]fs.FileInfo, error) {
	disk, err := diskfs.Open(p.parent.diskDevice)
	if err != nil {
		return nil, err
	}

	fs, err := disk.GetFilesystem(p.Index + 1)
	if err != nil {
		return nil, err
	}

	return fs.ReadDir(path)
}

func (p *Partition) OpenFile(path string, flag int) (io.ReadWriteCloser, error) {
	disk, err := diskfs.Open(p.parent.diskDevice)
	if err != nil {
		return nil, err
	}

	fs, err := disk.GetFilesystem(p.Index + 1)
	if err != nil {
		return nil, err
	}

	return fs.OpenFile(path, flag)
}

func (p *Partition) OpenStream() (io.ReadWriteCloser, error) {
	file, err := os.OpenFile(p.parent.diskDevice, os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}

	startOffset := int64(p.parent.partitionTable.SectorSize) * int64(p.Start)
	n, err := file.Seek(startOffset, 0)
	if n != startOffset {
		return nil, fmt.Errorf("couldn't seek: %d != %d", n, startOffset)
	}

	return file, err
}
