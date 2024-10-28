package main

import (
	"bytes"
	"log"
	"net"
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

func main() {
	privKey := crypto.GeneratePrivateKey()
	localNode := makeServer("LOCAL", &privKey, ":3000", []string{":4000"})
	go localNode.Start()

	remoteNode := makeServer("REMOTE_NODE", nil, ":4000", []string{":5000"})
	go remoteNode.Start()

	remoteNodeB := makeServer("REMOTE_NODE_B", nil, ":5000", nil)
	go remoteNodeB.Start()
	// tr := network.NewTCPTransport(":3000")
	// go tr.Start()

	go func() {
		time.Sleep(16 * time.Second)

		lateNode := makeServer("LATE_NODE", nil, ":6000", []string{":4000"})
		go lateNode.Start()
	}()

	time.Sleep(time.Second)

	tcpTest()
	//
	select {}
}

func makeServer(id string, pk *crypto.PrivateKey, addr string, seedNodes []string) *network.Server {
	opts := network.ServerOpts{
		SeedNodes:  seedNodes,
		ListenAddr: addr,
		PrivateKey: pk,
		ID:         id,
	}

	s, err := network.NewServer(opts)
	if err != nil {
		log.Fatal(err)
	}

	return s
}

func tcpTest() {
	conn, err := net.Dial("tcp", ":3000")
	if err != nil {
		panic(err)
	}

	privKey := crypto.GeneratePrivateKey()
	data := []byte{0x01, 0x0a, 0x02, 0x0a, 0x0b}
	tx := core.NewTransaction(data)
	tx.Sign(privKey)
	buf := &bytes.Buffer{}

	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		panic(err)
	}

	msg := network.NewMessage(network.MessageTypeTx, buf.Bytes())

	_, err = conn.Write(msg.Bytes())
	if err != nil {
		panic(err)
	}

}

// var transports = []network.Transport{
//     network.NewLocalTransport("LOCAL"),
//     // network.NewLocalTransport("REMOTE_B"),
//     // network.NewLocalTransport("REMOTE_C"),
// }

// func main() {
//     initRemoteServers(transports)
//     localNode := transports[0]
//     trLate := network.NewLocalTransport("LATE_NODE")
//     // remoteNodeA := transports[1]
//     // remoteNodeC := transports[3]
//
//     // go func() {
//     // 	for {
//     // 		if err := sendTransaction(remoteNodeA, localNode.Addr()); err != nil {
//     // 			logrus.Error(err)
//     // 		}
//     // 		time.Sleep(2 * time.Second)
//     // 	}
//     // }()
//
//     go func() {
//         time.Sleep(7 * time.Second)
//         // connect the late node with localNode
//         lateServer := makeLateServer(string(trLate.Addr()), trLate, nil)
//         go lateServer.Start()
//     }()
//
//     privKey := crypto.GeneratePrivateKey()
//     localServer := makeServer("LOCAL", localNode, &privKey)
//     localServer.Start()
// }
//
// func initRemoteServers(trs []network.Transport) {
//     for i := 0; i < len(trs); i++ {
//         id := fmt.Sprintf("REMOTE_%d", i)
//         s := makeServer(id, trs[i], nil)
//         go s.Start()
//     }
// }
//
//
// // WA
// func makeLateServer(id string, tr network.Transport, pk *crypto.PrivateKey) *network.Server {
//     opts := network.ServerOpts{
//         Transport:  tr,
//         PrivateKey: pk,
//         ID:         id,
//         Transports: append(transports, tr),
//     }
//
//     s, err := network.NewServer(opts)
//     if err != nil {
//         log.Fatal(err)
//     }
//
//     return s
// }
//
// func sendGetStatusMessage(tr network.Transport, to network.NetAddr) error {
//     var (
//         getStatusMsg = new(network.GetStatusMessage)
//         buf          = new(bytes.Buffer)
//     )
//     if err := gob.NewEncoder(buf).Encode(getStatusMsg); err != nil {
//         return err
//     }
//     msg := network.NewMessage(network.MessageTypeGetStatus, buf.Bytes())
//     return tr.SendMessage(to, msg.Bytes())
// }
//
// func sendTransaction(tr network.Transport, to network.NetAddr) error {
//     privKey := crypto.GeneratePrivateKey()
//     data := []byte{0x01, 0x0a, 0x02, 0x0a, 0x0b}
//     tx := core.NewTransaction(data)
//     tx.Sign(privKey)
//     buf := &bytes.Buffer{}
//
//     if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
//         return err
//     }
//
//     msg := network.NewMessage(network.MessageTypeTx, buf.Bytes())
//     tr.SendMessage(to, msg.Bytes())
//
//     return nil
// }
