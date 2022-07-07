package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/client"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/server"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/version"
)

func main() {
	fmt.Println("Updater version ", version.VERSION)
	flag.Bool("client", false, "Run as client")

	if strings.HasPrefix(os.Args[0], "client") || containsClient(os.Args[1:]) {
		client.Main()
	} else {
		server.Main()
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
