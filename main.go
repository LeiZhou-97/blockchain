package main

import (
	"fmt"
	"time"

	"github.com/LeiZhou-97/blockchain/network"
)

// Server
// Transport => tcp, udp,
// Block
// TX
// Keypair

func main() {
	fmt.Println("hello, world")

	trlocal := network.NewLocalTransport("LOCAL")
	trRemote := network.NewLocalTransport("REMOTE")

	trlocal.Connect(trRemote)
	trRemote.Connect(trlocal)

	go func() {
		for {
			trRemote.SendMessage(trlocal.Addr(), []byte("hello!"))
			time.Sleep(2 * time.Second)
		}
	}()

	opts := network.ServerOpts {
		Transports: []network.Transport{trlocal},
	}

	s := network.NewServer(opts)
	s.Start()
}