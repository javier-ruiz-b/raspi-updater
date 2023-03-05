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
}

func NewUpdater(conf *config.ClientConfig) *Updater {
	return &Updater{
		conf: conf,
	}
}

func (u *Updater) Run() error {
	qc := transport.NewQuicClient(*u.conf.Address, *u.conf.Log)
	selfupdater := selfupdater.NewSelfUpdater(qc, u.conf.Runner)
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
	disk := disk.NewDisk(*u.conf.DiskDevice)
	err = disk.Read()
	if err != nil {
		return err
	}

	pr.SetDescription("Reading local version", 2)
	localVersion, err := disk.ReadVersion()
	if err != nil {
		pr.Printf("Warning: could not read local version: %v", err)
	}

	pr.SetDescription("Checking for image update for "+*u.conf.Id, 3)
	imageUrlVersion := strings.Replace(server.API_IMAGES_VERSION, "{id}", *u.conf.Id, 1)
	serverVersion, err := qc.GetString(imageUrlVersion)
	if err != nil {
		return err
	}

	if localVersion == serverVersion {
		pr.Printf("Up to date!")
		return nil
	}

	pr.Printf("Update Available. Server version: %s, Client version:%s", serverVersion, localVersion)
	err = u.update(qc, disk, pr)

	if err != nil {
		return err
	}

	pr.Printf("Update success!")
	return nil
}

func (u *Updater) update(qc transport.Client, myDisk *disk.Disk, pr progress.Progress) error {
	pr.SetDescription("Getting partition scheme", 5)
	imageUrlVersion := strings.Replace(server.API_IMAGES_PARTITION_TABLE, "{id}", *u.conf.Id, 1)
	var remotePartitionTable disk.PartitionTable
	if err := qc.GetObject(imageUrlVersion, &remotePartitionTable); err != nil {
		return err
	}

	remoteBootPartition, err := remotePartitionTable.GetBootPartition()
	if err != nil {
		return err
	}

	pr.SetDescription("Downloading boot partition", 6)
	compressedBootPartitionInTemp, err := os.CreateTemp(os.TempDir(), "boot")
	if err != nil {
		return err
	}

	downloadUrlBoot := strings.Replace(server.API_IMAGES_DOWNLOAD, "{id}", *u.conf.Id, 1)
	downloadUrlBoot = strings.Replace(downloadUrlBoot, "{partitionIndex}", strconv.Itoa(remoteBootPartition.Index), 1)
	downloadUrlBoot = strings.Replace(downloadUrlBoot, "{compression}", *u.conf.CompressionTool, 1)
	if err := qc.DownloadFile(compressedBootPartitionInTemp.Name(), downloadUrlBoot, progress.NewProgressReporter(pr, 20)); err != nil {
		return err
	}

	pr.SetDescription("Checking boot partition", 20)
	if err := compression.CheckFile(*u.conf.CompressionTool, compressedBootPartitionInTemp.Name()); err != nil {
		return err
	}

	pr.SetDescription("Partitioning disk if necessary", 21)
	if err := myDisk.MergePartitionTable(&remotePartitionTable); err != nil {
		return err
	}

	pr.Printf("Partition table:\nRemote: %s\nFinal: %s",
		remotePartitionTable.GetInfo(),
		myDisk.GetPartitionTable().GetInfo())
	if err := myDisk.Write(); err != nil {
		return err
	}

	pr.SetDescription("Rereading partition table", 21)
	if err := u.conf.Runner.RunPath("/bin/partprobe"); err != nil {
		return err
	}
	if err := myDisk.Read(); err != nil {
		return err
	}

	pr.SetDescription("Writing boot partition", 22)
	localBootPatition, err := myDisk.GetPartitionTable().GetBootPartition()
	if err != nil {
		return err
	}
	if localBootPatition.Size != remoteBootPartition.Size {
		return fmt.Errorf("local and remote boot partition sizes differ. Local: %d != Remote: %d",
			localBootPatition.Size, remoteBootPartition.Size)
	}

	bootSize := int64(remoteBootPartition.Size)
	counter := progress.NewIoCounter(bootSize, progress.NewProgressReporter(pr, 30))

	bootPartitionStream, err := localBootPatition.OpenStream()
	if err != nil {
		return err
	}

	compressedBootPartitionInTemp.Seek(0, 0)
	compressor := compression.NewStreamDecompressor(io.TeeReader(compressedBootPartitionInTemp, counter), bootPartitionStream, *u.conf.CompressionTool)
	if err = compressor.Run(); err != nil {
		return err
	}

	return nil
}
