package client

import (
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/config"
)

func ClientMain() error {
	conf := config.NewClientConfig()
	conf.LoadFlags()

	return RunClient(conf)
}

func RunClient(conf *config.ClientConfig) error {
	return NewUpdater(conf).Run()
}
