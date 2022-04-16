package protocol

import "encoding/gob"

func Send(enc *gob.Encoder, item packet) error {
	if err := enc.Encode(item.packetType()); err != nil {
		return err
	}
	return enc.Encode(item)
}
