package server

import (
	"flag"
	"fmt"
	"log"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/version"
)

func Main(port int) {
	fmt.Println("Server", version.VERSION)
	var ()
	flag.Parse()

	log.Print("Port: ", port)

}
