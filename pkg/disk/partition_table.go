package disk

import (
	"fmt"

	"github.com/dustin/go-humanize"
)

type PartitionTable struct {
	Partitions []Partition
	Size       uint64
	SectorSize int
}

func (p *PartitionTable) Print() {
	fmt.Printf("Partition table TotalSize %s, NumSectors %d, SectorSize %d\n", humanize.Bytes(p.Size), p.Size, p.SectorSize)
	for i := 0; i < len(p.Partitions); i++ {
		partition := p.Partitions[i]
		fmt.Printf(" - Partition type 0x%02x,  TotalSize %7s,  Start %8d\n", partition.Type, humanize.Bytes(uint64(partition.Size)*uint64(p.SectorSize)), partition.Start)
	}
}
