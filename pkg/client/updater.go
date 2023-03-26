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

	err = u.backupDisk(u.backupLengthLocal(u.disk.GetPartitionTable()), false)
	if err != nil {
		return err
	}

	log.Print("Checking for image update for " + *u.conf.Id)
	imageUrlVersion := strings.Replace(server.API_IMAGES_VERSION, "{id}", *u.conf.Id, 1)
	serverVersion, err := u.qc.GetString(imageUrlVersion)
	if err != nil {
		log.Print("No image found for ", *u.conf.Id)
		return nil
	}

	if localVersion == serverVersion {
		log.Print("Up to date.")
		return nil
	}

	log.Printf("Image update available.\n  Server version: %s\n  Client version: %s\n", serverVersion, localVersion)
	if err = u.update(); err != nil {
		return err
	}

	if err = u.disk.Read(); err != nil {
		log.Printf("Couldn't reread disk")
	}

	if err = u.disk.WriteVersion(serverVersion); err != nil {
		log.Printf("Couldn't write version %s after update", serverVersion)
	}

	if err := u.conf.Runner.RunPath("/bin/sync"); err != nil {
		log.Printf("Couldn't sync")
	}

	log.Printf("Update complete. Rebooting in 5 seconds")
	u.conf.Runner.RunPath("/usr/bin/sleep", "5")
	if err := u.conf.Runner.RunPath("/usr/bin/busybox", "reboot", "-f"); err != nil {
		log.Printf("Couldn't reboot")
	}

	return nil
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
	downloadedBootPartitionPath, err := u.downloadBootPartition(u.partitionDownloadUrl(remoteBootPartition))
	if err != nil {
		return err
	}
	defer os.Remove(downloadedBootPartitionPath)

	localPartitionInfo := u.disk.GetPartitionTable().GetInfo()
	if err := u.backupDisk(u.backupLengthLocalAndRemote(u.disk.GetPartitionTable(), &remotePartitionTable), true); err != nil {
		return err
	}

	log.Print("Merging partition table")
	if err := u.disk.MergePartitionTable(&remotePartitionTable); err != nil {
		return err
	}
	mergedPartitionTable := u.disk.GetPartitionTable()

	log.Printf("Partition table:\nLocal: %s\nRemote: %s\nFinal: %s",
		localPartitionInfo,
		remotePartitionTable.GetInfo(),
		mergedPartitionTable.GetInfo())
	log.Print("Writing partition table")
	if err := u.disk.Write(); err != nil {
		return err
	}

	log.Print("Running partprobe")
	if err := u.conf.Runner.RunPath("/bin/partprobe"); err != nil {
		return err
	}

	log.Print("Reading disk")
	if err := u.disk.Read(); err != nil {
		return err
	}

	log.Print("Writing boot partition")
	localBootPartition, err := u.disk.GetPartitionTable().GetBootPartition()
	if err != nil {
		return err
	}
	compressedBootPartition, err := os.Open(downloadedBootPartitionPath)
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

func (u *Updater) downloadBootPartition(url string) (string, error) {
	compressedBootPartition, err := os.CreateTemp(os.TempDir(), "boot")
	if err != nil {
		return "", err
	}
	defer compressedBootPartition.Close()

	// Test compression on the fly
	streamTestReader, streamTestWriter := io.Pipe()
	tester := compression.NewStreamTester(streamTestReader, *u.conf.CompressionTool)
	if err := tester.Open(); err != nil {
		return "", err
	}
	defer tester.Close()

	downloadStream, contentLength, err := u.qc.GetDownloadStream(url)
	if err != nil {
		return "", err
	}

	bar := progressbar.DefaultBytes(contentLength, "Download boot partition")
	defer bar.Close()

	// Copy the file data to the output file
	buffer := make([]byte, 1*1024*1024) // 1 MB
	if _, err := io.CopyBuffer(io.MultiWriter(compressedBootPartition, streamTestWriter, bar), downloadStream, buffer); err != nil {
		return "", err
	}

	if err = streamTestWriter.Close(); err != nil {
		return "", err
	}

	if _, err = tester.Read(make([]byte, 1)); err != io.EOF && err != nil {
		return "", fmt.Errorf("failed testing compressed stream: %s", err.Error())
	}

	return compressedBootPartition.Name(), nil
}

func (u *Updater) backupDisk(backupDiskLength int64, force bool) error {
	log.Print("Backing up disk if necessary")
	localVersion, _ := u.disk.ReadVersion()
	if localVersion == "" {
		localVersion = "0"
	}

	if !force {
		backupExists, _ := u.backupExists(localVersion)
		if backupExists {
			log.Println("Backup exists on server.")
			return nil
		}
	}

	disk, err := os.Open(*u.conf.DiskDevice)
	if err != nil {
		return err
	}
	defer disk.Close()

	diskBarReader := progressbar.DefaultBytes(backupDiskLength, "Backup "+*u.conf.DiskDevice)
	defer diskBarReader.Close()

	diskCompressionStream := compression.NewStreamCompressorN(io.TeeReader(disk, diskBarReader), backupDiskLength, *u.conf.CompressionTool)
	if err := diskCompressionStream.Open(); err != nil {
		return err
	}
	defer diskCompressionStream.Close()

	backupUrl := strings.Replace(server.API_IMAGES_BACKUP, "{id}", *u.conf.Id, 1)
	backupUrl = strings.Replace(backupUrl, "{version}", localVersion, 1)
	backupUrl = strings.Replace(backupUrl, "{compression}", *u.conf.CompressionTool, 1)

	return u.qc.UploadStream(backupUrl, diskCompressionStream)
}

func (u *Updater) backupExists(localVersion string) (bool, error) {
	backupExistsUrl := strings.Replace(server.API_IMAGES_BACKUP_EXISTS, "{id}", *u.conf.Id, 1)
	backupExistsUrl = strings.Replace(backupExistsUrl, "{version}", localVersion, 1)
	var backupExists bool
	err := u.qc.GetObject(backupExistsUrl, backupExists)
	if err != nil {
		log.Println("Error requesting backup status: ", err)
		return false, err
	}
	return backupExists, nil
}

func (u *Updater) backupLengthLocal(localPartitionTable *disk.PartitionTable) int64 {
	if len(localPartitionTable.Partitions) == 0 {
		return int64(localPartitionTable.Size)
	}
	lastLocalPartition := localPartitionTable.Partitions[len(localPartitionTable.Partitions)-1]
	return int64(lastLocalPartition.EndSector()) * int64(localPartitionTable.SectorSize)
}

func (u *Updater) backupLengthLocalAndRemote(localPartitionTable, remotePartitionTable *disk.PartitionTable) int64 {
	lastRemotePartition := remotePartitionTable.Partitions[len(remotePartitionTable.Partitions)-1]
	lastRemoteSector := lastRemotePartition.EndSector()
	for _, localPartition := range localPartitionTable.Partitions {
		if localPartition.EndSector() >= lastRemoteSector {
			return int64(localPartition.EndSector()) * int64(localPartitionTable.SectorSize)
		}
	}
	return int64(localPartitionTable.Size)
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
	defer bar.Close()
	buffer := make([]byte, 1*1024*1024)
	if _, err := io.CopyBuffer(io.MultiWriter(partitionStream, bar), remoteBootPartitionStream, buffer); err != nil {
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
