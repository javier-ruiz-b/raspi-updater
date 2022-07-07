package client

import (
	"errors"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/config"
)

func Main() {
	conf := config.NewClientConfig()
	conf.LoadFlags()

	RunClient(conf)
}

func RunClient(conf *config.ClientConfig) error {
	qc := NewQuicClient(*conf.Address, *conf.Log)
	err := update(qc, *conf.Version)
	if err != nil {
		return err
	}

	return errors.New("not implemented")
}
