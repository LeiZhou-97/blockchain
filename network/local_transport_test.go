package network

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnect(t *testing.T) {
	tra := NewLocalTransport("A")
	trb := NewLocalTransport("B")

	tra.Connect(trb)
	trb.Connect(tra)
	assert.Equal(t, tra.peers[trb.addr], trb)
}


func TestSendMessage(t *testing.T) {
	tra := NewLocalTransport("A")
	trb := NewLocalTransport("B")

	tra.Connect(trb)
	trb.Connect(tra)

	assert.Nil(t, tra.SendMessage(trb.addr, []byte("hello world")))

	rpc := <- trb.Consume()
	assert.Equal(t, rpc.Payload, []byte("hello world"))
	assert.Equal(t, rpc.From, tra.addr)
}