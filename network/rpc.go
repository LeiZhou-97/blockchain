package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"

	"github.com/LeiZhou-97/blockchain/core"
	"github.com/sirupsen/logrus"
)

type MessageType byte

const (
	MessageTypeTx MessageType = 0x1
	MessageTypeBlock 
)

type RPC struct {
	From 	NetAddr
	Payload io.Reader
}

type Message struct {
	Header MessageType
	Data []byte
}

func NewMessage(t MessageType, data []byte) *Message {
	return &Message{
		Header: t,
		Data: data,
	}
}

func (msg *Message) Bytes() []byte {
	buf := &bytes.Buffer{}
	gob.NewEncoder(buf).Encode(msg)
	return buf.Bytes()
}

type DecodeMessage struct {
	From NetAddr
	Data any
}

type RPCDecodeFunc func(RPC) (*DecodeMessage, error)

func DefaultRPCDecodeFunc(rpc RPC) (*DecodeMessage, error) {
	msg := Message{}
	if err := gob.NewDecoder(rpc.Payload).Decode(&msg); err != nil {
		return nil, fmt.Errorf("failed to decode message from %s: %s", rpc.From, err)
	}

	logrus.WithFields(logrus.Fields{
		"from": rpc.From,
		"type": msg.Header,
	}).Debug(" incoming msg ")

	switch msg.Header {
		case MessageTypeTx:
			tx := new(core.Transaction)
			if err := tx.Decode(core.NewGobTxDecoder(bytes.NewReader(msg.Data))); err != nil {
				return nil, err
			}
			return &DecodeMessage{
				From: rpc.From,
				Data: tx,
			}, nil
		default:
			return nil, fmt.Errorf("invalid message header %x", msg.Header)
	}
}


type RPCProcessor interface {
	ProcessTransaction(*DecodeMessage) error
}
