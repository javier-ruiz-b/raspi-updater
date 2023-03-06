package client

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/compression"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/config"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/disk"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/progress"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/selfupdater"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/server"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/transport"
)

type Updater struct {
	conf *config.ClientConfig
	disk *disk.Disk
	qc   transport.Client
}

func NewUpdater(conf *config.ClientConfig) *Updater {
	return &Updater{
		conf: conf,
	}
}

func (u *Updater) Run() error {
	u.qc = transport.NewQuicClient(*u.conf.Address, *u.conf.Log)
	selfupdater := selfupdater.NewSelfUpdater(u.qc, u.conf.Runner)
	pr := progress.NewMainProgressReporter()

	pr.SetDescription("Checking for client update", 0)
	updateAvailable, err := selfupdater.IsUpdateAvailable(*u.conf.Version)
	if err != nil {
		return err
	}

	if updateAvailable {
		pr.SetDescription("Update found", 25)
		return selfupdater.DownloadAndRunUpdate(progress.NewProgressReporter(pr, 100))
	}

	pr.SetDescription("Reading local disk", 1)
	_, err = os.Stat(*u.conf.DiskDevice)
	if err != nil {
		return err
	}

	u.disk = disk.NewDisk(*u.conf.DiskDevice)
	if err = u.disk.Read(); err != nil {
		return err
	}

	pr.SetDescription("Reading local version", 2)
	localVersion, err := u.disk.ReadVersion()
	if err != nil {
		pr.Printf("Warning: could not read local version: %v", err)
	}

	pr.SetDescription("Checking for image update for "+*u.conf.Id, 3)
	imageUrlVersion := strings.Replace(server.API_IMAGES_VERSION, "{id}", *u.conf.Id, 1)
	serverVersion, err := u.qc.GetString(imageUrlVersion)
	if err != nil {
		return err
	}

	if localVersion == serverVersion {
		pr.Printf("Up to date!")
		return nil
	}

	pr.Printf("Update Available. Server version: %s, Client version:%s", serverVersion, localVersion)
	return u.update(pr)
}

func (u *Updater) update(pr progress.Progress) error {
	pr.SetDescription("Getting partition scheme", 5)
	imageUrlVersion := strings.Replace(server.API_IMAGES_PARTITION_TABLE, "{id}", *u.conf.Id, 1)
	var remotePartitionTable disk.PartitionTable
	if err := u.qc.GetObject(imageUrlVersion, &remotePartitionTable); err != nil {
		return err
	}

	remoteBootPartition, err := remotePartitionTable.GetBootPartition()
	if err != nil {
		return err
	}

	pr.SetDescription("Downloading boot partition", 6)
	compressedBootPartition, err := os.CreateTemp(os.TempDir(), "boot")
	if err != nil {
		return err
	}
	defer os.Remove(compressedBootPartition.Name())

	if err := u.qc.DownloadFile(compressedBootPartition.Name(), u.partitionDownloadUrl(remoteBootPartition), progress.NewProgressReporter(pr, 20)); err != nil {
		return err
	}

	pr.SetDescription("Checking downloaded boot partition", 20)
	if err := compression.CheckFile(*u.conf.CompressionTool, compressedBootPartition.Name()); err != nil {
		return err
	}
	compressedBootPartitionInfo, err := os.Stat(compressedBootPartition.Name())
	if err != nil {
		return err
	}
	localPartitionInfo := u.disk.GetPartitionTable().GetInfo()
	pr.SetDescription("Partitioning disk if necessary", 21)
	if err := u.disk.MergePartitionTable(&remotePartitionTable); err != nil {
		return err
	}

	pr.Printf("Partition table:\nLocal: %s\nRemote: %s\nFinal: %s",
		localPartitionInfo,
		remotePartitionTable.GetInfo(),
		u.disk.GetPartitionTable().GetInfo())
	if err := u.disk.Write(); err != nil {
		return err
	}

	pr.SetDescription("Rereading partition table", 21)
	if err := u.conf.Runner.RunPath("/bin/partprobe"); err != nil {
		return err
	}
	if err := u.disk.Read(); err != nil {
		return err
	}

	pr.SetDescription("Writing boot partition", 22)
	localBootPartition, err := u.disk.GetPartitionTable().GetBootPartition()
	if err != nil {
		return err
	}

	bootPartitionStream, err := localBootPartition.OpenStream()
	if err != nil {
		return err
	}

	compressedBootPartition.Seek(0, 0)
	counter := progress.NewIoCounter(compressedBootPartitionInfo.Size(), progress.NewProgressReporter(pr, 30))
	compressor := compression.NewStreamDecompressor(bootPartitionStream, io.TeeReader(compressedBootPartition, counter), *u.conf.CompressionTool)
	if err = compressor.Run(); err != nil {
		return err
	}
	compressedBootPartition.Close()
	os.Remove(compressedBootPartition.Name())

	pr.SetDescription("Syncing disks", 31)
	if err := u.conf.Runner.RunPath("/bin/sync"); err != nil {
		return err
	}
	for i, partition := range remotePartitionTable.Partitions {
		if i == remoteBootPartition.Index {
			continue
		}
		localPartition, err := u.disk.GetPartitionTable().GetBootPartition()
		if err != nil {
			return err
		}
		localPartitionStream, err := localPartition.OpenStream()
		if err != nil {
			return err
		}
		defer localPartitionStream.Close()

		minPercent := i * 100 / len(remotePartitionTable.Partitions)
		maxPercent := (i + 1) * 100 / len(remotePartitionTable.Partitions)
		pr.SetDescription(fmt.Sprintf("Downloading partition %d / %d", i+1, len(remotePartitionTable.Partitions)), minPercent)

		downloadStream, _, err := u.qc.GetDownloadStream(u.partitionDownloadUrl(&partition))
		if err != nil {
			return err
		}
		defer downloadStream.Close()

		partitionSizeBytes := int64(partition.Size * uint32(u.disk.GetPartitionTable().SectorSize))
		counter = progress.NewIoCounter(partitionSizeBytes, progress.NewProgressReporter(pr, maxPercent))
		compressor := compression.NewStreamDecompressor(io.MultiWriter(localPartitionStream, counter), downloadStream, *u.conf.CompressionTool)
		if err := compressor.Run(); err != nil {
			return err
		}
	}

	pr.SetDescription("Syncing disks", 100)

	return u.conf.Runner.RunPath("/bin/sync")
}

func (u *Updater) partitionDownloadUrl(partition *disk.Partition) string {
	downloadUrlBoot := strings.Replace(server.API_IMAGES_DOWNLOAD, "{id}", *u.conf.Id, 1)
	downloadUrlBoot = strings.Replace(downloadUrlBoot, "{partitionIndex}", strconv.Itoa(partition.Index), 1)
	downloadUrlBoot = strings.Replace(downloadUrlBoot, "{compression}", *u.conf.CompressionTool, 1)
	return downloadUrlBoot
}
