package config

import (
	"flag"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/version"
)

type Config struct {
	Address          *string
	CertificatesPath *string
	Verbose          *bool
	Log              *bool
	Version          *string
}

func NewConfig() *Config {
	defaultAddress := "localhost:31416"
	defaultCertsPath := ""
	defaultVerbose := false
	defaultLog := false
	defaultVersion := version.VERSION

	return &Config{
		Address:          &defaultAddress,
		CertificatesPath: &defaultCertsPath,
		Verbose:          &defaultVerbose,
		Log:              &defaultLog,
		Version:          &defaultVersion,
	}
}

func (c *Config) LoadFlags() {
	c.Address = flag.String("address", *c.Address, "Server address")
	c.CertificatesPath = flag.String("certs", *c.CertificatesPath, "Additional certificates path")
	c.Verbose = flag.Bool("verbose", *c.Verbose, "Increase verbosity")
	c.Log = flag.Bool("log", *c.Log, "Log quic communication in qlog files")
}
