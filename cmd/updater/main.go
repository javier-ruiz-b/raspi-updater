package main

import (
	"flag"
	"os"
	"strings"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/client"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/server"
)

func main() {
	var (
		_    = flag.Bool("client", false, "Run as client")
		port = flag.Int("port", 31416, "Updater TCP port")
	)

	if strings.HasPrefix(os.Args[0], "client") || containsClient(os.Args[1:]) {
		client.Main(*port)
	} else {
		server.Main(*port)
	}
}

func containsClient(args []string) bool {
	for _, value := range args {
		if (value) == "-client" {
			return true
		}
	}
	return false
}
