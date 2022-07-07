package client

import (
	"errors"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/config"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/transport"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/updater"
)

func ClientMain() error {
	conf := config.NewClientConfig()
	conf.LoadFlags()

	return RunClient(conf, &updater.OsRunner{})
}

func RunClient(conf *config.ClientConfig, runner updater.Runner) error {
	qc := transport.NewQuicClient(*conf.Address, *conf.Log)
	updater := updater.NewUpdater(qc, runner)

	updateAvailable, err := updater.IsUpdateAvailable(*conf.Version)
	if err != nil {
		return err
	}
	if updateAvailable {
		return updater.DownloadAndRunUpdate()
	}

	return errors.New("not implemented")
}
