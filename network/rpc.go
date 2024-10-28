package network

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io"
	"net"

	"github.com/LeiZhou-97/blockchain/core"
	"github.com/sirupsen/logrus"
)

type MessageType byte

const (
	MessageTypeTx MessageType = 0x1
	MessageTypeBlock MessageType = 0x2
	MessageTypeGetBlocks MessageType = 0x3
	MessageTypeStatus MessageType = 0x4
	MessageTypeGetStatus MessageType = 0x5
	MessageTypeBlocks MessageType = 0x6
)

type RPC struct {
	From 	net.Addr
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
	From net.Addr
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
		case MessageTypeGetBlocks:
			getBlocks := new(GetBlocksMessage)
			if err := gob.NewDecoder(bytes.NewReader(msg.Data)).Decode(getBlocks); err != nil {
				return nil, err
			}
			return &DecodeMessage{
				From: rpc.From,
				Data: getBlocks,
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
		case MessageTypeBlocks:
			blocks := new(BlocksMessage)
			if err := gob.NewDecoder(bytes.NewReader(msg.Data)).Decode(blocks); err != nil {
				return nil, err
			}
			return &DecodeMessage{
				From: rpc.From,
				Data: blocks,
			}, nil
		default:
			return nil, fmt.Errorf("invalid message header %x", msg.Header)
	}
}


type RPCProcessor interface {
	ProcessMessage(*DecodeMessage) error
}

func init() {
	gob.Register(elliptic.P256())
}
