package client

import (
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/version"
)

func Main(port int) {
	fmt.Println("Client", version.VERSION)
	var (
		address    = flag.String("address", "", "Server address (e.g. test.com)")
		id         = flag.String("id", "", "Client ID (e.g. rpi_john_garage)")
		diskDevice = flag.String("disk", "/dev/mmcblk0", "Disk device")
	)
	flag.Parse()

	log.Print("")
	log.Print("Address: ", *address)
	log.Print("TCP port: ", port)

	//disk.Create("a")
	Client(&Config{
		ServerAddress: *address,
		Id:            *id,
		DiskDevice:    *diskDevice,
	})
}

func Client(config *Config) error {
	return errors.New("Not implemented")
}
