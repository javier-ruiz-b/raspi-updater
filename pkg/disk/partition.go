package disk

import (
	"io"
	"io/fs"

	diskfs "github.com/diskfs/go-diskfs"
)

type Partition struct {
	Type   PartitionType
	Start  uint32 // in megabyte
	Size   uint32 // in megabyte
	index  int
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

func (p *Partition) ReadDir(path string) ([]fs.FileInfo, error) {
	disk, err := diskfs.Open(p.parent.diskDevice)
	if err != nil {
		return nil, err
	}

	fs, err := disk.GetFilesystem(p.index + 1)
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

	fs, err := disk.GetFilesystem(p.index + 1)
	if err != nil {
		return nil, err
	}

	return fs.OpenFile(path, flag)
}
