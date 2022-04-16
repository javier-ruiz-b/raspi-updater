package parser

import (
	"encoding/gob"
	"fmt"
	"io"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/nlog"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/protocol"
)

type Parser struct {
	id             string
	lastPacketType protocol.PacketType
	Enc            *gob.Encoder
	Dec            *gob.Decoder
	hello          *protocol.Hello
}

func NewParser(id string, enc *gob.Encoder, dec *gob.Decoder) *Parser {
	return &Parser{
		id:    id,
		Enc:   enc,
		Dec:   dec,
		hello: protocol.NewHello(id),
	}
}

func (p *Parser) Parse() error {
	switch p.lastPacketType {
	case protocol.PacketHello:
		return p.parseHello()
	case protocol.PacketBye:
		nlog.Debug("received bye packet")
		return io.EOF
	default:
		return fmt.Errorf("unknown packet type %d", p.lastPacketType)
	}
}

func (p *Parser) parseHello() error {
	var hello protocol.Hello
	err := p.Dec.Decode(&hello)
	if err != nil {
		return err
	}

	return protocol.Send(p.Enc, p.hello)
}

func (p *Parser) ReadPacketType() (protocol.PacketType, error) {
	err := p.Dec.Decode(&p.lastPacketType)
	return p.lastPacketType, err
}
