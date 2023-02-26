package client

import (
	"errors"
	"fmt"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/config"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/disk"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/progress"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/selfupdater"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/transport"
)

type Updater struct {
	conf *config.ClientConfig
	disk *disk.Disk
}

func NewUpdater(conf *config.ClientConfig) *Updater {
	return &Updater{
		conf: conf,
		disk: disk.NewDisk(*conf.DiskDevice),
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
	err = u.disk.Read()
	if err != nil {
		return err
	}
	u.disk.GetPartitionTable().Print()

	pr.SetDescription("Reading local version", 2)

	pr.SetDescription("Checking for image update", 2)
	imageUrl := fmt.Sprintf("/images/%s", *u.conf.Id)
	_, err = qc.GetString(imageUrl + "/version")
	if err != nil {
		return err
	}

	return errors.New("not implemented")
}
