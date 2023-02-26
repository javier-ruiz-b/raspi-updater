package disk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreatesExt4PartitionIfDoesNotExist(t *testing.T) {
	desiredTable := PartitionTable{
		Size: 1024,
		Partitions: []Partition{
			{
				Type:  Fat32CHS,
				Start: 0,
				Size:  128,
			},
			{
				Type:  Linux,
				Start: 128,
				Size:  1024 - 128,
			},
		},
	}

	existingTable := PartitionTable{
		Size: 1024,
		Partitions: []Partition{
			{
				Type:  Fat32CHS,
				Start: 0,
				Size:  128,
			},
		},
	}

	result, err := mergePartitionTables(&desiredTable, &existingTable)

	assert.Nil(t, err)

	//we create the desired table
	assert.Equal(t, &desiredTable, result)
}

func TestFailsIfBootPartitionIsMissingOnDesiredTable(t *testing.T) {
	desiredTable := PartitionTable{
		Size: 1024,
		Partitions: []Partition{
			{
				Type:  Linux,
				Start: 128,
				Size:  1024 - 128,
			},
		},
	}

	existingTable := PartitionTable{
		Size: 1024,
		Partitions: []Partition{
			{
				Type:  Fat32CHS,
				Start: 0,
				Size:  128,
			},
		},
	}

	_, err := mergePartitionTables(&desiredTable, &existingTable)

	assert.NotNil(t, err)
}

func TestFailsIfDesiredPartitionTableIsEmpty(t *testing.T) {
	desiredTable := PartitionTable{
		Size:       1024,
		Partitions: []Partition{},
	}
	existingTable := PartitionTable{
		Size: 1024,
		Partitions: []Partition{
			{
				Type:  Fat32CHS,
				Start: 0,
				Size:  128,
			},
		},
	}

	_, err := mergePartitionTables(&desiredTable, &existingTable)

	assert.NotNil(t, err)
}

func TestFailsIfDifferentSectorSize(t *testing.T) {
	desiredTable := PartitionTable{
		SectorSize: 4096,
	}
	existingTable := PartitionTable{
		SectorSize: 512,
	}

	_, err := mergePartitionTables(&desiredTable, &existingTable)

	assert.NotNil(t, err)
}

func TestCreatesBootPartitionIfMissing(t *testing.T) {
	desiredTable := PartitionTable{
		Size: 1024,
		Partitions: []Partition{
			{
				Type:  Fat32CHS,
				Start: 0,
				Size:  128,
			},
		},
	}
	existingTable := PartitionTable{
		Size:       1024,
		Partitions: []Partition{},
	}

	result, err := mergePartitionTables(&desiredTable, &existingTable)

	assert.Nil(t, err)
	assert.Equal(t, &desiredTable, result)
}

func TestDoesNotModifyBootPartitionIfDiffersFromDesiredPartitionTable(t *testing.T) {
	desiredTable := PartitionTable{
		Size: 1024,
		Partitions: []Partition{
			{
				Type:  Fat32CHS,
				Start: 16,
				Size:  128,
			},
		},
	}
	existingTable := PartitionTable{
		Size: 1024,
		Partitions: []Partition{
			{
				Type:  Fat32CHS,
				Start: 0,
				Size:  256,
			},
		},
	}

	result, err := mergePartitionTables(&desiredTable, &existingTable)

	assert.Nil(t, err)
	assert.Equal(t, &existingTable, result)
}

func TestDoesNotOverlapOnExistingPartition(t *testing.T) {
	desiredTable := PartitionTable{
		Size: 1024,
		Partitions: []Partition{
			{
				Type:  Fat32CHS,
				Start: 0,
				Size:  128,
			},
			{
				Type:  Linux,
				Start: 128,
				Size:  128,
			},
		},
	}
	existingTable := PartitionTable{
		Size: 1024,
		Partitions: []Partition{
			{
				Type:  Fat32CHS,
				Start: 0,
				Size:  256,
			},
		},
	}

	result, err := mergePartitionTables(&desiredTable, &existingTable)

	assert.Nil(t, err)
	assert.Equal(t, &PartitionTable{
		Size: 1024,
		Partitions: []Partition{
			{
				Type:  Fat32CHS,
				Start: 0,
				Size:  256,
			},
			{
				Type:  Linux,
				Start: 256,
				Size:  128,
			},
		},
	}, result)
}

func TestFailsIfItPartitionsDoNotFit(t *testing.T) {
	desiredTable := PartitionTable{
		Size: 2048,
		Partitions: []Partition{
			{
				Type:  Fat32CHS,
				Start: 0,
				Size:  512,
			},
			{
				Type:  Linux,
				Start: 512,
				Size:  1024,
			},
		},
	}
	existingTable := PartitionTable{
		Size: 2048,
		Partitions: []Partition{
			{
				Type:  Fat32CHS,
				Start: 0,
				Size:  1536,
			},
		},
	}

	_, err := mergePartitionTables(&desiredTable, &existingTable)

	assert.NotNil(t, err)
}

func TestFailsIfDesiredDiskSizeDoesNotFit(t *testing.T) {
	desiredTable := PartitionTable{
		Size:       2048,
		Partitions: []Partition{},
	}
	existingTable := PartitionTable{
		Size:       1024,
		Partitions: []Partition{},
	}

	_, err := mergePartitionTables(&desiredTable, &existingTable)

	assert.NotNil(t, err)
}
