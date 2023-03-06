package disk

import (
	"errors"
	"fmt"

	"github.com/dustin/go-humanize"
)

func mergePartitionTables(desired *PartitionTable, existing *PartitionTable) (*PartitionTable, error) {
	if len(desired.Partitions) == 0 {
		return nil, errors.New("the desired partition table is empty")
	}

	if desired.SectorSize != existing.SectorSize {
		return nil, fmt.Errorf("sector size is not equal. Desired sector size: %d, Existing sector size: %d",
			desired.SectorSize, existing.SectorSize)
	}

	if desired.Size > existing.Size {
		return nil, fmt.Errorf("the resulting partition table does not fit in the disk. Required: %s. Available: %s",
			humanize.Bytes(desired.Size), humanize.Bytes(existing.Size))
	}

	desiredBoot, err := desired.GetBootPartition()
	if err != nil {
		return nil, err
	}

	partitions := []Partition{}

	if len(existing.Partitions) == 0 || !isFatPartition(&existing.Partitions[0]) {
		partitions = append(partitions, *desiredBoot)
	} else {
		partitions = append(partitions, existing.Partitions[0])
	}

	totalSize := uint64(partitions[0].Size)

	for i := 1; i < len(desired.Partitions); i++ {
		desiredPartition := desired.Partitions[i]
		previousPartition := partitions[i-1]
		newPartition := Partition{
			Type:  desiredPartition.Type,
			Start: previousPartition.Start + previousPartition.Size,
			Size:  desiredPartition.Size,
		}

		partitions = append(partitions, newPartition)
		totalSize += uint64(newPartition.Size)
	}

	if totalSize > existing.Size {
		return nil, fmt.Errorf("the resulting partition table does not fit in the disk. Required: %s, available: %s",
			humanize.Bytes(totalSize), humanize.Bytes(existing.Size))
	}

	return &PartitionTable{
		Size:       desired.Size,
		SectorSize: desired.SectorSize,
		Partitions: partitions,
	}, nil
}

func isFatPartition(part *Partition) bool {
	switch part.Type {
	case Fat12, Fat16, Fat16b, Fat16bLBA, Fat32CHS, Fat32LBA:
		return true
	}
	return false
}
