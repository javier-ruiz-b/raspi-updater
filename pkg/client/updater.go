package client

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/compression"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/config"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/disk"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/selfupdater"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/server"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/transport"
	"github.com/schollz/progressbar/v3"
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
	u.qc = transport.NewQuicClient(u.conf)
	selfupdater := selfupdater.NewSelfUpdater(u.qc, u.conf.Runner)

	log.Print("Checking for client update")
	updateAvailable, err := selfupdater.IsUpdateAvailable(*u.conf.Version)
	if err != nil {
		return err
	}

	if updateAvailable {
		log.Print("Update found")
		return selfupdater.DownloadAndRunUpdate()
	}

	log.Print("Reading local disk")
	_, err = os.Stat(*u.conf.DiskDevice)
	if err != nil {
		return err
	}

	u.disk = disk.NewDisk(*u.conf.DiskDevice)
	if err = u.disk.Read(); err != nil {
		return err
	}

	log.Print("Reading local version")
	localVersion, err := u.disk.ReadVersion()
	if err != nil {
		log.Printf("Warning: could not read local version: %v", err)
	}

	log.Print("Checking for image update for " + *u.conf.Id)
	imageUrlVersion := strings.Replace(server.API_IMAGES_VERSION, "{id}", *u.conf.Id, 1)
	serverVersion, err := u.qc.GetString(imageUrlVersion)
	if err != nil {
		return err
	}

	if localVersion == serverVersion {
		log.Print("Up to date!")
		return nil
	}

	log.Printf("Update Available. Server version: %s, Client version:%s", serverVersion, localVersion)
	return u.update()
}

func (u *Updater) update() error {
	log.Print("Getting partition scheme")
	imageUrlVersion := strings.Replace(server.API_IMAGES_PARTITION_TABLE, "{id}", *u.conf.Id, 1)
	var remotePartitionTable disk.PartitionTable
	if err := u.qc.GetObject(imageUrlVersion, &remotePartitionTable); err != nil {
		return err
	}

	remoteBootPartition, err := remotePartitionTable.GetBootPartition()
	if err != nil {
		return err
	}

	log.Print("Downloading boot partition")
	compressedBootPartition, err := os.CreateTemp(os.TempDir(), "boot")
	if err != nil {
		return err
	}
	defer os.Remove(compressedBootPartition.Name())

	if err := u.qc.DownloadFile(compressedBootPartition.Name(), u.partitionDownloadUrl(remoteBootPartition)); err != nil {
		return err
	}
	if _, err = compressedBootPartition.Seek(0, 0); err != nil {
		return err
	}

	fmt.Print("Checking downloaded boot partition")
	if err := compression.CheckFile(*u.conf.CompressionTool, compressedBootPartition.Name()); err != nil {
		return err
	}

	localPartitionInfo := u.disk.GetPartitionTable().GetInfo()
	log.Print("Partitioning disk if necessary")
	if err := u.disk.MergePartitionTable(&remotePartitionTable); err != nil {
		return err
	}

	log.Printf("Partition table:\nLocal: %s\nRemote: %s\nFinal: %s",
		localPartitionInfo,
		remotePartitionTable.GetInfo(),
		u.disk.GetPartitionTable().GetInfo())
	if err := u.disk.Write(); err != nil {
		return err
	}
	log.Print("Rereading partition table")
	if err := u.conf.Runner.RunPath("/bin/partprobe"); err != nil {
		return err
	}
	if err := u.disk.Read(); err != nil {
		return err
	}

	log.Print("Writing boot partition")
	localBootPartition, err := u.disk.GetPartitionTable().GetBootPartition()
	if err != nil {
		return err
	}

	if err = u.writePartition(localBootPartition, compressedBootPartition, "Writing boot partition"); err != nil {
		return err
	}

	if err = compressedBootPartition.Close(); err != nil {
		return err
	}
	if err = os.Remove(compressedBootPartition.Name()); err != nil {
		return err
	}

	for i, partition := range remotePartitionTable.Partitions {
		if i == remoteBootPartition.Index {
			continue
		}
		localPartition := &u.disk.GetPartitionTable().Partitions[i]
		downloadStream, _, err := u.qc.GetDownloadStream(u.partitionDownloadUrl(&partition))
		if err != nil {
			return err
		}
		defer downloadStream.Close()

		progressText := fmt.Sprintf("Transferring partition %d / %d", i+1, len(remotePartitionTable.Partitions))
		if err = u.writePartition(localPartition, downloadStream, progressText); err != nil {
			return err
		}
	}

	return nil
}

func (u *Updater) writePartition(partition *disk.Partition, compressedInput io.ReadCloser, progressText string) error {
	partitionStream, err := partition.OpenStream()
	if err != nil {
		return err
	}

	remoteBootPartitionStream := compression.NewStreamDecompressor(compressedInput, *u.conf.CompressionTool)
	if err := remoteBootPartitionStream.Open(); err != nil {
		return err
	}

	bar := progressbar.DefaultBytes(int64(partition.SizeBytes()), progressText)
	if _, err := io.Copy(io.MultiWriter(partitionStream, bar), remoteBootPartitionStream); err != nil {
		return err
	}

	if err = remoteBootPartitionStream.Close(); err != nil {
		return err
	}

	if err = partitionStream.Close(); err != nil {
		return err
	}

	if err := u.conf.Runner.RunPath("/bin/sync"); err != nil {
		return err
	}

	return nil
}
func (u *Updater) partitionDownloadUrl(partition *disk.Partition) string {
	downloadUrlBoot := strings.Replace(server.API_IMAGES_DOWNLOAD, "{id}", *u.conf.Id, 1)
	downloadUrlBoot = strings.Replace(downloadUrlBoot, "{partitionIndex}", strconv.Itoa(partition.Index), 1)
	downloadUrlBoot = strings.Replace(downloadUrlBoot, "{compression}", *u.conf.CompressionTool, 1)
	return downloadUrlBoot
}
