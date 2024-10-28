package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"time"

	"github.com/LeiZhou-97/blockchain/core"
	"github.com/LeiZhou-97/blockchain/crypto"
	"github.com/LeiZhou-97/blockchain/network"
)

// Server
// Transport => tcp, udp,
// Block
// TX
// Keypair

var transports = []network.Transport{
	network.NewLocalTransport("LOCAL"),
	// network.NewLocalTransport("REMOTE_B"),
	// network.NewLocalTransport("REMOTE_C"),
}

func main() {
	initRemoteServers(transports)
	localNode := transports[0]
	trLate := network.NewLocalTransport("LATE_NODE")
	// remoteNodeA := transports[1]
	// remoteNodeC := transports[3]

	// go func() {
	// 	for {
	// 		if err := sendTransaction(remoteNodeA, localNode.Addr()); err != nil {
	// 			logrus.Error(err)
	// 		}
	// 		time.Sleep(2 * time.Second)
	// 	}
	// }()

	go func() {
		time.Sleep(7 * time.Second)
		// connect the late node with localNode
		lateServer := makeLateServer(string(trLate.Addr()), trLate, nil)
		go lateServer.Start()
	}()

	privKey := crypto.GeneratePrivateKey()
	localServer := makeServer("LOCAL", localNode, &privKey)
	localServer.Start()
}

func initRemoteServers(trs []network.Transport) {
	for i := 0; i < len(trs); i++ {
		id := fmt.Sprintf("REMOTE_%d", i)
		s := makeServer(id, trs[i], nil)
		go s.Start()
	}
}

func makeServer(id string, tr network.Transport, pk *crypto.PrivateKey) *network.Server {
	opts := network.ServerOpts{
		Transport:  tr,
		PrivateKey: pk,
		ID:         id,
		Transports: transports,
	}

	s, err := network.NewServer(opts)
	if err != nil {
		log.Fatal(err)
	}

	return s
}

// WA
func makeLateServer(id string, tr network.Transport, pk *crypto.PrivateKey) *network.Server {
	opts := network.ServerOpts{
		Transport:  tr,
		PrivateKey: pk,
		ID:         id,
		Transports: append(transports, tr),
	}

	s, err := network.NewServer(opts)
	if err != nil {
		log.Fatal(err)
	}

	return s
}

func sendGetStatusMessage(tr network.Transport, to network.NetAddr) error {
	var (
		getStatusMsg = new(network.GetStatusMessage)
		buf          = new(bytes.Buffer)
	)
	if err := gob.NewEncoder(buf).Encode(getStatusMsg); err != nil {
		return err
	}
	msg := network.NewMessage(network.MessageTypeGetStatus, buf.Bytes())
	return tr.SendMessage(to, msg.Bytes())
}

func sendTransaction(tr network.Transport, to network.NetAddr) error {
	privKey := crypto.GeneratePrivateKey()
	data := []byte{0x01, 0x0a, 0x02, 0x0a, 0x0b}
	tx := core.NewTransaction(data)
	tx.Sign(privKey)
	buf := &bytes.Buffer{}

	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}

	msg := network.NewMessage(network.MessageTypeTx, buf.Bytes())
	tr.SendMessage(to, msg.Bytes())

	return nil
}
