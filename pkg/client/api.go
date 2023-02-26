package client

import "github.com/javier-ruiz-b/raspi-image-updater/pkg/transport"

type Api struct {
	c transport.Client
}

func NewApi(c transport.Client) *Api {
	return &Api{
		c: c,
	}
}
