package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/LeiZhou-97/blockchain/api"
	"github.com/LeiZhou-97/blockchain/core"
	"github.com/LeiZhou-97/blockchain/crypto"
	"github.com/LeiZhou-97/blockchain/types"
	"github.com/go-kit/log"
)

var defaultBlockTime = 5 * time.Second

type ServerOpts struct {
	APIListenAddr    string
	SeedNodes     []string
	ListenAddr    string
	TCPTransport  *TCPTransport
	ID            string
	Logger        log.Logger
	RPCDecodeFunc RPCDecodeFunc
	RPCProcessor  RPCProcessor
	BlockTIme     time.Duration
	PrivateKey    *crypto.PrivateKey
}

type Server struct {
	ServerOpts
	TCPTransport *TCPTransport
	mu           sync.RWMutex
	peerCh       chan *TCPPeer
	peerMap      map[net.Addr]*TCPPeer
	mempool      *TxPool
	chain        *core.BlockChain
	isValidator  bool
	rpcCh        chan RPC
	quitCh       chan struct{}
}

func NewServer(opts ServerOpts) (*Server, error) {
	if opts.BlockTIme == time.Duration(0) {
		opts.BlockTIme = defaultBlockTime
	}
	if opts.RPCDecodeFunc == nil {
		opts.RPCDecodeFunc = DefaultRPCDecodeFunc
	}
	if opts.Logger == nil {
		opts.Logger = log.NewLogfmtLogger(os.Stderr)
		opts.Logger = log.With(opts.Logger, "addr", opts.ID)
	}

	chain, err := core.NewBlockChain(opts.Logger, genesisBlock())
	if err != nil {
		return nil, err
	}

	if opts.APIListenAddr != "" {
		apiServercfg := api.ServerConfig{
			Logger: opts.Logger,
			ListenAddr: opts.APIListenAddr,
		}

		apiServer := api.NewServer(apiServercfg, chain)

		go apiServer.Start()

		opts.Logger.Log("msg", "json api server running", "port", opts.APIListenAddr)
	}

	peerCh := make(chan *TCPPeer)
	tr := NewTCPTransport(opts.ListenAddr, peerCh)

	s := &Server{
		ServerOpts:   opts,
		TCPTransport: tr,
		peerCh:       peerCh,
		peerMap:      make(map[net.Addr]*TCPPeer),
		mempool:      NewTxPool(1000),
		chain:        chain,
		isValidator:  opts.PrivateKey != nil,
		rpcCh:        make(chan RPC),
		quitCh:       make(chan struct{}, 1),
	}

	s.TCPTransport.peerCh = peerCh

	// if we do not get any processor form the server opts, we going to
	// use the server as default
	if s.RPCProcessor == nil {
		s.RPCProcessor = s
	}

	if s.isValidator {
		go s.validatorLoop()
	}

	return s, nil
}

func (s *Server) bootstrapNetwork() {
	for _, addr := range s.SeedNodes {
		fmt.Println("trying to connect to: ", addr)
		go func(addr string) {
			conn, err := net.Dial("tcp", addr)

			if err != nil {
				fmt.Printf("could not connect to %+v\n", conn)
				return
			}

			s.peerCh <- &TCPPeer{
				conn: conn,
			}
		}(addr)

	}
}

func (s *Server) Start() {
	s.TCPTransport.Start()

	s.Logger.Log("msg", "accepting TCP connection on", "addr", s.ListenAddr, "id", s.ID)
	s.bootstrapNetwork()

free:
	for {
		select {
		case peer := <-s.peerCh:
			s.peerMap[peer.conn.RemoteAddr()] = peer
			go peer.readLoop(s.rpcCh)

			if err := s.sendGetStatusMessage(peer); err != nil {
				s.Logger.Log("err:", err)
				continue
			}

			s.Logger.Log("msg", "peer added to the server", "Outgoing", peer.Outgoing, "addr", peer.conn.RemoteAddr())
		// consumer
		case rpc := <-s.rpcCh:
			msg, err := s.RPCDecodeFunc(rpc)
			if err != nil {
				s.Logger.Log("error", err)
				continue
			}
			if err := s.RPCProcessor.ProcessMessage(msg); err != nil {
				if err != core.ErrBlockKnown {
					s.Logger.Log("error", err)
				}
			}
		case <-s.quitCh:
			break free
		}
	}

	s.Logger.Log("msg", "Server shutdown")
}

func (s *Server) validatorLoop() {
	s.Logger.Log("msg", "Starting validatorLoop")
	ticker := time.NewTicker(s.BlockTIme)

	for {
		<-ticker.C
		if err := s.createNewBlock(); err != nil {
			s.Logger.Log("err", err)
		}
	}
}

func (s *Server) ProcessMessage(dmsg *DecodeMessage) error {
	switch t := dmsg.Data.(type) {
	case *core.Transaction:
		return s.processTransaction(t)
	case *core.Block:
		return s.processBlock(t)
	case *GetStatusMessage:
		return s.processGetStatusMessage(dmsg.From, t)
	case *StatusMessage:
		return s.processStatusMessage(dmsg.From, t)
	case *GetBlocksMessage:
		return s.processGetBlocksMessage(dmsg.From, t)
	case *BlocksMessage:
		return s.processBlocksMessage(dmsg.From, t)
	}
	return nil
}

func (s *Server) processGetBlocksMessage(from net.Addr, data *GetBlocksMessage) error {
	fmt.Printf("received getBlocksMessage => %+v\n", data)	

	blocks := []*core.Block{}
	
	ourHeight := s.chain.Height()
	if data.To == 0 {
		for i:=int(data.From); i<=int(ourHeight); i++ {
			blcok, err := s.chain.GetBlock(uint32(i))
			if err != nil {
				return err
			}

			blocks = append(blocks, blcok)
		}
	}

	blocksMsg := &BlocksMessage{
		Blocks: blocks,
	}
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(blocksMsg); err != nil {
		return err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	msg := NewMessage(MessageTypeBlocks, buf.Bytes())
	peer, ok := s.peerMap[from]
	if !ok {
		return fmt.Errorf("peer %s not known", peer.conn.RemoteAddr())
	}
	return peer.Send(msg.Bytes())
}

func (s *Server) sendGetStatusMessage(peer *TCPPeer) error {
	var (
		getStatusMsg = new(GetStatusMessage)
		buf          = new(bytes.Buffer)
	)
	if err := gob.NewEncoder(buf).Encode(getStatusMsg); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeGetStatus, buf.Bytes())

	/// ask to sync with peer
	return peer.Send(msg.Bytes())
}

func (s *Server) broadcast(payload []byte) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for netAddr, peer := range s.peerMap {
		if err := peer.Send(payload); err != nil {
			fmt.Printf("peer send err => addr %s [err : %s]", netAddr, err)
		}
	}
	return nil
}

func (s *Server) processBlocksMessage(from net.Addr, data *BlocksMessage) error {
	s.Logger.Log("msg", "received BLOCKS!!!!!!!!", "from", from)

	for _, block := range data.Blocks {
		fmt.Printf("BlOCK with %+v\n", block.Header)
		if err := s.chain.AddBlock(block); err != nil {
			fmt.Errorf("late node err add block: %s\n", err)
			return err
		}
	}

	return nil
}

func (s *Server) processStatusMessage(from net.Addr, data *StatusMessage) error {
	fmt.Printf("=> received status msg from %s => %+v\n", from, data)
	if data.CurrentHeight <= s.chain.Height() {
		s.Logger.Log("msg", "cannot sync blockHeight to low", "ourHeight", s.chain.Height(), "theirHeight", data.CurrentHeight, "addr", from)
		return nil
	}

	go s.requestBlocksLoop(from)
	return nil
}

func (s *Server) processGetStatusMessage(from net.Addr, data *GetStatusMessage) error {
	s.Logger.Log("msg", "received get status msg", "from", from)

	statusMessage := &StatusMessage{
		CurrentHeight: s.chain.Height(),
		ID:            s.ID,
	}
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(statusMessage); err != nil {
		return err
	}

	// response status msg
	s.mu.RLock()
	defer s.mu.RUnlock()
	msg := NewMessage(MessageTypeStatus, buf.Bytes())
	peer, ok := s.peerMap[from]
	if !ok {
		return fmt.Errorf("peer %s not known", peer.conn.RemoteAddr())
	}
	return peer.Send(msg.Bytes())
}

// TODO: Find a way to make sure we dont keep syncing when we are at the highest
// block height in the network.
func (s *Server) requestBlocksLoop(peer net.Addr) error {
	ticker := time.NewTicker(3 * time.Second)
	for {
		ourHeight := s.chain.Height()
		s.Logger.Log("msg", "requesting new blocks", "requesting height", ourHeight+1)
		// In this case we are 100% sure that the node has blocks heigher than us.
		getBlocksMessage := &GetBlocksMessage{
			From: ourHeight + 1,
			To:   0,
		}
		buf := new(bytes.Buffer)
		if err := gob.NewEncoder(buf).Encode(getBlocksMessage); err != nil {
			return err
		}
		s.mu.RLock()
		defer s.mu.RUnlock()
		msg := NewMessage(MessageTypeGetBlocks, buf.Bytes())
		peer, ok := s.peerMap[peer]
		if !ok {
			return fmt.Errorf("peer %s not known", peer.conn.RemoteAddr())
		}
		if err := peer.Send(msg.Bytes()); err != nil {
			s.Logger.Log("error", "failed to send to peer", "err", err, "peer", peer)
		}
		<-ticker.C
	}
}

func (s *Server) processBlock(b *core.Block) error {
	if err := s.chain.AddBlock(b); err != nil {
		return err
	}
	go s.broadcastBlock(b)

	return nil
}

func (s *Server) processTransaction(tx *core.Transaction) error {
	hash := tx.Hash(core.TxHasher{})

	if s.mempool.Contains(hash) {
		return nil
	}

	if err := tx.Verify(); err != nil {
		return err
	}

	s.Logger.Log("msg", "adding new tx to mempool",
		"hash", hash,
		"mempoolLen", s.mempool.PendingCount())

	//// TODO: broadcast this tx to peers <23-10-24, Lei Zhou> //
	go s.broadcastTx(tx)

	s.mempool.Add(tx)

	return nil
}

func (s *Server) broadcastBlock(b *core.Block) error {
	buf := &bytes.Buffer{}
	if err := b.Encode(core.NewGobBlockEncoder(buf)); err != nil {
		return err
	}

	msg := NewMessage(MessageTypeBlock, buf.Bytes())

	return s.broadcast(msg.Bytes())
}

func (s *Server) broadcastTx(tx *core.Transaction) error {
	buf := &bytes.Buffer{}

	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeTx, buf.Bytes())

	return s.broadcast(msg.Bytes())
}

func (s *Server) createNewBlock() error {
	currentHeader, err := s.chain.GetHeader(s.chain.Height())
	if err != nil {
		return err
	}

	// we are going to use all transactions that are in the pending memPool
	// Later on when we know the internal structure of our transaction
	// we ill implement some kind of complexity function to determine
	// how many transactions can be included in a block.
	txx := s.mempool.Pending()

	block, err := core.NewBlockFromPrevHeader(currentHeader, txx)
	if err != nil {
		return err
	}

	if err := block.Sign(*s.PrivateKey); err != nil {
		return err
	}

	if err := s.chain.AddBlock(block); err != nil {
		return err
	}

	s.mempool.ClearPending()

	go s.broadcastBlock(block)

	return nil
}

// func (s *Server) initTransport() {
//     for _, tr := range s.Transports {
//         go func(tr Transport) {
//             // block if no new element
//             for rpc := range tr.Consume() {
//                 s.rpcCh <- rpc
//             }
//         }(tr)
//     }
// }

func genesisBlock() *core.Block {
	header := &core.Header{
		Version:   1,
		DataHash:  types.Hash{},
		Height:    0,
		Timestamp: 000000,
	}
	b, _ := core.NewBlock(header, nil)
	privKey := crypto.GeneratePrivateKey()
	if err := b.Sign(privKey); err != nil {
		panic(err)
	}
	return b
}
