package protocol

import (
	"fmt"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/version"
)

type Hello struct {
	VersionMajor uint16
	VersionMinor uint16
	VersionPatch uint16
	Id           string
}

func NewHello(id string) *Hello {
	return &Hello{
		VersionMajor: version.MAJOR,
		VersionMinor: version.MINOR,
		VersionPatch: version.PATCH,
		Id:           id,
	}
}

func (*Hello) packetType() PacketType {
	return PacketHello
}

func (h *Hello) Version() string {
	return fmt.Sprintf("%d.%d.%d", h.VersionMajor, h.VersionMinor, h.VersionPatch)
}
