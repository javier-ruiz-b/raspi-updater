package disk

type PartitionTable struct {
	Partitions []Partition
	Size       uint64
	SectorSize int
}
