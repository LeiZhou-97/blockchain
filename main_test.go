package main

import (
	"testing"
	"time"

	"github.com/LeiZhou-97/blockchain/crypto"
	"github.com/LeiZhou-97/blockchain/network"
)

func TestMain(t *testing.T) {
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
		lateServer := makeServer(string(trLate.Addr()), trLate, nil)
		go lateServer.Start()
	}()

	privKey := crypto.GeneratePrivateKey()
	localServer := makeServer("LOCAL", localNode, &privKey)
	localServer.Start()
}