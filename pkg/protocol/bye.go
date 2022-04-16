package protocol

type Bye struct {
	Message string
}

func NewBye(message string) *Bye {
	return &Bye{
		Message: message,
	}
}

func (*Bye) packetType() PacketType {
	return PacketBye
}
