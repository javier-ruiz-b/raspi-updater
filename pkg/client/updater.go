package client

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"strings"

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
	pr := progress.NewProgressReporter()

	pr.SetDescription("Checking for client update", 0)
	updateAvailable, err := selfupdater.IsUpdateAvailable(*u.conf.Version)
	if err != nil {
		return err
	}

	if updateAvailable {
		pr.SetDescription("Update found", 25)
		return selfupdater.DownloadAndRunUpdate(progress.NewSubProgressReporter(pr, 100))
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

	disk.GetPartitionTable().Print()
	fmt.Printf("\n\n")

	pr.SetDescription("Reading local version", 2)
	localVersion, err := disk.ReadVersion()
	if err != nil {
		log.Printf("warning: could not read local version: %v", err)
	}

	pr.SetDescription("Checking for image update for "+*u.conf.Id, 3)
	imageUrlVersion := strings.Replace(server.API_IMAGES_VERSION, "{image}", *u.conf.Id, 1)
	serverVersion, err := qc.GetString(imageUrlVersion)
	if err != nil {
		return err
	}

	if localVersion == serverVersion {
		log.Printf("\nUp to date!")
		return nil
	}

	log.Printf("\nUpdate Available. Server version: %s, Client version:%s\n", serverVersion, localVersion)
	err = u.update(qc, disk, pr)

	if err != nil {
		return err
	}

	log.Printf("\nUpdate success!")
	return nil
}

func (u *Updater) update(qc transport.Client, myDisk *disk.Disk, pr progress.Progress) error {
	pr.SetDescription("Getting partition scheme", 5)
	imageUrlVersion := strings.Replace(server.API_IMAGES_PARTITION_TABLE, "{image}", *u.conf.Id, 1)
	partitionTableBytes, err := qc.GetBytes(imageUrlVersion)
	if err != nil {
		return err
	}

	dec := gob.NewDecoder(bytes.NewBuffer(partitionTableBytes))
	var remoteDisk disk.Disk
	err = dec.Decode(&remoteDisk)
	remoteDisk.GetPartitionTable().Print()

	return err
}
