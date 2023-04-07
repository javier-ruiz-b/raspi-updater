package config

import (
	"flag"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/version"
)

type Config struct {
	Address         *string
	CertificatePath *string
	KeyPath         *string
	Verbose         *bool
	Log             *bool
	Version         *string
	IsClient        *bool
}

func NewConfig() *Config {
	defaultAddress := "localhost:31416"
	defaultCertPath := ""
	defaultKeyPath := ""
	defaultVerbose := false
	defaultLog := false
	defaultVersion := version.VERSION
	defaultIsClient := false

	return &Config{
		Address:         &defaultAddress,
		CertificatePath: &defaultCertPath,
		KeyPath:         &defaultKeyPath,
		Verbose:         &defaultVerbose,
		Log:             &defaultLog,
		Version:         &defaultVersion,
		IsClient:        &defaultIsClient,
	}
}

func (c *Config) LoadFlags() {
	c.Address = flag.String("address", *c.Address, "Server address")
	c.CertificatePath = flag.String("certFile", *c.CertificatePath, "QUIC certificate file path")
	c.KeyPath = flag.String("keyFile", *c.KeyPath, "QUIC key file path")
	c.Verbose = flag.Bool("verbose", *c.Verbose, "Increase verbosity")
	c.Log = flag.Bool("log", *c.Log, "Log quic communication in qlog files")
	c.IsClient = flag.Bool("client", *c.IsClient, "Run as client")
}
