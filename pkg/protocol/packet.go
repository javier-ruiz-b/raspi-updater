package protocol

type PacketType byte

type packet interface {
	packetType() PacketType
}

// List of GUID partition types
const (
	PacketInvalid        PacketType = 0x00
	PacketHello          PacketType = 0x01
	PacketBye            PacketType = 0x02
	PacketFileInfo       PacketType = 0x03
	PacketFileData       PacketType = 0x04
	PacketPartitionTable PacketType = 0x05
)
