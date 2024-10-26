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
	MessageTypeBlock MessageType = 0x2
	MessageTypeStatus MessageType = 0x3
	MessageTypeGetStatus MessageType = 0x4
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
		case MessageTypeBlock:
			b := new(core.Block)	
			if err := b.Decode(core.NewGobBlockDecoder(bytes.NewReader(msg.Data))); err != nil {
				return nil, err
			}
			return &DecodeMessage{
				From: rpc.From,
				Data: b,
			}, nil
		case MessageTypeGetStatus:
			return &DecodeMessage{
				From: rpc.From,
				Data: &GetStatusMessage{},
			}, nil
		case MessageTypeStatus:
			statusMessage := new(StatusMessage)
			if err := gob.NewDecoder(bytes.NewReader(msg.Data)).Decode(statusMessage); err != nil {
				return nil, err
			}
			return &DecodeMessage{
				From: rpc.From,
				Data: statusMessage,	
			}, nil
		default:
			return nil, fmt.Errorf("invalid message header %x", msg.Header)
	}
}


type RPCProcessor interface {
	ProcessMessage(*DecodeMessage) error
}
