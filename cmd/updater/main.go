package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/client"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/server"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/version"
)

func main() {
	fmt.Println("Updater version ", version.VERSION)

	go signalCatcher()

	var err error
	if strings.HasPrefix(os.Args[0], "client") || containsClient(os.Args[1:]) {
		err = client.ClientMain()
	} else {
		err = server.ServerMain()
	}

	if err != io.EOF && err != nil {
		log.Fatal("Error: ", err)
	}

	os.Exit(0)
}

func containsClient(args []string) bool {
	for _, value := range args {
		if value == "-client" || value == "--client" {
			return true
		}
	}
	return false
}

func signalCatcher() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Fatal("Received SIGTERM")
	}()
}
