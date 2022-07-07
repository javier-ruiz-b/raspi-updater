package server

import (
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/config"
)

func ServerMain() error {
	options := config.NewServerConfig()
	options.LoadFlags()

	server := NewServer(options)
	return server.Listen()
}
