package server

import (
	"io"
	"log"
	"os"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/config"
)

func Main() {
	options := config.NewServerConfig()
	options.LoadFlags()

	server := NewServer(options)
	err := server.Listen()

	if err != io.EOF && err != nil {
		log.Print("Server error: ", err)
		os.Exit(1)
	}
}
