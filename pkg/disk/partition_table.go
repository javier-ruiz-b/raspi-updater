package disk

import (
	"fmt"

	"github.com/dustin/go-humanize"
)

type PartitionTable struct {
	Partitions []Partition
	Size       uint64 // in bytes
	SectorSize int    // in bytes
}

func (p *PartitionTable) GetInfo() string {
	result := fmt.Sprintf("Partition table TotalSize %s, SectorSize %d\n", humanize.Bytes(p.Size), p.SectorSize)
	for i := 0; i < len(p.Partitions); i++ {
		partition := p.Partitions[i]
		result = result + fmt.Sprintf(" - Partition type 0x%02x,  TotalSize %7s,  Start %8d\n", partition.Type, humanize.Bytes(uint64(partition.Size)*uint64(p.SectorSize)), partition.Start)
	}
	return result
}

func (p *PartitionTable) GetBootPartition() (*Partition, error) {
	if len(p.Partitions) == 0 {
		return nil, fmt.Errorf("no partitions found")
	}

	bootPartition := &p.Partitions[0]
	if !isFatPartition(bootPartition) {
		return nil, fmt.Errorf("first partition (/boot) of the desired partition table is supposed to be a FAT filesystem. Found %b on %d with size %d",
			bootPartition.Type, bootPartition.Start, bootPartition.Size)
	}

	return bootPartition, nil
}
