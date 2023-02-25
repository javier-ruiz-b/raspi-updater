package client

import (
	"errors"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/config"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/progress"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/selfupdater"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/transport"
)

type Updater struct {
	conf *config.ClientConfig
}

func NewUpdater(conf *config.ClientConfig) *Updater {
	return &Updater{conf: conf}
}

func (u *Updater) Run() error {
	qc := transport.NewQuicClient(*u.conf.Address, *u.conf.Log)
	selfupdater := selfupdater.NewSelfUpdater(qc, u.conf.Runner)
	pr := progress.NewProgressReporter()

	pr.SetDescription("Checking for client updates", 0)
	updateAvailable, err := selfupdater.IsUpdateAvailable(*u.conf.Version)
	if err != nil {
		return err
	}

	if updateAvailable {
		pr.SetDescription("Update found", 25)
		return selfupdater.DownloadAndRunUpdate(progress.NewSubProgressReporter(pr, 100))
	}

	return errors.New("not implemented")
}
