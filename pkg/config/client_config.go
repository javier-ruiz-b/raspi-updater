package config

import (
	"flag"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/runner"
)

type ClientConfig struct {
	*Config
	Id              *string
	DiskDevice      *string
	CompressionTool *string
	Runner          runner.Runner
}

func NewClientConfig() *ClientConfig {
	defaultId := ""
	defaultDisk := "/dev/mmcblk0"

	result := &ClientConfig{
		Config:     NewConfig(),
		Id:         &defaultId,
		DiskDevice: &defaultDisk,
		Runner:     &runner.OsRunner{},
	}

	return result
}

func (c *ClientConfig) LoadFlags() {
	c.Config.LoadFlags()

	c.Id = flag.String("id", *c.Id, "Client ID (e.g. rpi_john_garage)")
	c.DiskDevice = flag.String("disk", *c.DiskDevice, "Disk device")
	c.CompressionTool = flag.String("compression", *c.CompressionTool, "Compression format for transfers (xz, lz4, gzip, etc..)")

	flag.Parse()
}
