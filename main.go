package main

import (
	"bytes"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/LeiZhou-97/blockchain/core"
	"github.com/LeiZhou-97/blockchain/crypto"
	"github.com/LeiZhou-97/blockchain/network"
	"github.com/sirupsen/logrus"
)

// Server
// Transport => tcp, udp,
// Block
// TX
// Keypair

func main() {

	trlocal := network.NewLocalTransport("LOCAL")
	trRemote := network.NewLocalTransport("REMOTE")
	// trc := network.NewLocalTransport("C")

	trlocal.Connect(trRemote)
	trRemote.Connect(trlocal)

	go func() {
		for {
			if err := sendTransaction(trRemote, trlocal.Addr()); err != nil {
				logrus.Error(err)
			}
			time.Sleep(2 * time.Second)
		}
	}()

	privKey := crypto.GeneratePrivateKey()
	opts := network.ServerOpts{
		PrivateKey: &privKey,
		ID: "LOCAL",
		Transports: []network.Transport{trlocal},
	}

	s, err := network.NewServer(opts)
	if err != nil {
		log.Fatal(err)
	}
	s.Start()
}

func sendTransaction(tr network.Transport, to network.NetAddr) error {
	privKey := crypto.GeneratePrivateKey()
	data := []byte(strconv.FormatInt(int64(rand.Intn(1000)), 10))
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
