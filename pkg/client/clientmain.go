package client

import (
	"errors"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/config"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/selfupdater"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/transport"
)

func ClientMain() error {
	conf := config.NewClientConfig()
	conf.LoadFlags()

	return RunClient(conf, &selfupdater.OsRunner{})
}

func RunClient(conf *config.ClientConfig, runner selfupdater.Runner) error {
	qc := transport.NewQuicClient(*conf.Address, *conf.Log)
	selfupdater := selfupdater.NewSelfUpdater(qc, runner)

	updateAvailable, err := selfupdater.IsUpdateAvailable(*conf.Version)
	if err != nil {
		return err
	}
	if updateAvailable {
		return selfupdater.DownloadAndRunUpdate()
	}

	return errors.New("not implemented")
}
