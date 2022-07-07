package config

import "flag"

type ServerConfig struct {
	*Config
	ImagesDir  *string
	UpdaterDir *string
}

func NewServerConfig() *ServerConfig {
	defaultImagesDir := "./images"
	defaultUpdaterDir := "./bin"
	result := &ServerConfig{
		Config:     NewConfig(),
		ImagesDir:  &defaultImagesDir,
		UpdaterDir: &defaultUpdaterDir,
	}
	return result
}

func (c *ServerConfig) LoadFlags() {
	c.Config.LoadFlags()

	c.ImagesDir = flag.String("images", *c.ImagesDir, "Images directory")
	c.UpdaterDir = flag.String("updater", *c.UpdaterDir, "Updater binaries directory")

	flag.Parse()
}
