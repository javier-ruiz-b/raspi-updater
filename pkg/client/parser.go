package client

import (
	"encoding/gob"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/parser"
)

type ClientParser struct {
	*parser.Parser
}

func NewClientParser(id string, enc *gob.Encoder, dec *gob.Decoder) *ClientParser {
	return &ClientParser{
		Parser: parser.NewParser(id, enc, dec),
	}
}

func (p *ClientParser) Parse() error {
	packet, err := p.ReadPacketType()
	if err != nil {
		return err
	}
	switch packet {
	default:
		return p.Parser.Parse()
	}
}
