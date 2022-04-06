package client

import (
	"flag"
	"fmt"
	"log"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/version"
)

func Main(port int) {
	fmt.Println("Client", version.VERSION)
	var (
		address = flag.String("address", "", "Server address (may include wireguard port)")
	)
	flag.Parse()

	log.Print("")
	log.Print("Address: ", *address)
	log.Print("TCP port: ", port)
}
