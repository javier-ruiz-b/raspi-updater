package client

import (
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/protocol"
	"github.com/stretchr/testify/assert"
)

func TestParsesHello(t *testing.T) {
	var clientToServer bytes.Buffer
	var serverToClient bytes.Buffer
	tested := NewClientParser("Client", gob.NewEncoder(&clientToServer), gob.NewDecoder(&serverToClient))

	//server sends Hello
	protocol.Send(gob.NewEncoder(&serverToClient), protocol.NewHello("Server"))

	assert.Nil(t, tested.Parse())

	clientToServerDec := gob.NewDecoder(&clientToServer)

	var packetType protocol.PacketType
	assert.Nil(t, clientToServerDec.Decode(&packetType))
	assert.Equal(t, protocol.PacketHello, packetType)

	var readHello protocol.Hello
	assert.Nil(t, clientToServerDec.Decode(&readHello))
	assert.Equal(t, "Client", readHello.Id)
}
