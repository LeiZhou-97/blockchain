package network

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/LeiZhou-97/blockchain/core"
	"github.com/LeiZhou-97/blockchain/crypto"
	"github.com/LeiZhou-97/blockchain/types"
	"github.com/go-kit/log"
	"github.com/sirupsen/logrus"
)

var defaultBlockTime = 5 * time.Second

type ServerOpts struct {
	ID            string
	Transport     Transport
	Logger        log.Logger
	RPCDecodeFunc RPCDecodeFunc
	RPCProcessor  RPCProcessor
	Transports    []Transport
	BlockTIme     time.Duration
	PrivateKey    *crypto.PrivateKey
}

type Server struct {
	ServerOpts
	mempool     *TxPool
	chain *core.BlockChain
	isValidator bool
	rpcCh       chan RPC
	quitCh      chan struct{}
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
		opts.Logger = log.With(opts.Logger, "ID", opts.ID)
	}

	chain, err := core.NewBlockChain(opts.Logger,genesisBlock())
	if err != nil {
		return nil, err
	}
	s := &Server{
		ServerOpts:  opts,
		mempool:     NewTxPool(1000),
		chain: chain,
		isValidator: opts.PrivateKey != nil,
		rpcCh:       make(chan RPC),
		quitCh:      make(chan struct{}, 1),
	}
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

func (s *Server) Start() {
	s.initTransport()

free:
	for {
		select {
		case rpc := <-s.rpcCh:
			msg, err := s.RPCDecodeFunc(rpc)
			if err != nil {
				s.Logger.Log("error",err)
			}
			if err := s.RPCProcessor.ProcessMessage(msg); err != nil {
				logrus.Error(err)
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
	}
	return nil
}

func (s *Server) broadcast(payload []byte) error {
	for _, tr := range s.Transports {
		if err := tr.Broadcast(payload); err != nil {
			return err
		}
	}
	return nil
}


func (s *Server) processStatusMessage(from NetAddr, msg *StatusMessage) error {
	fmt.Printf("received status msg from %s => %+v\n", from, msg)

	return nil
}

func (s *Server) processGetStatusMessage(from NetAddr, msg *GetStatusMessage) error {
	fmt.Printf("received status msg from %s => %+v\n", from, msg)
	return nil
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

func (s *Server) initTransport() {
	for _, tr := range s.Transports {
		go func(tr Transport) {
			// block if no new element
			for rpc := range tr.Consume() {
				s.rpcCh <- rpc
			}
		}(tr)
	}

}

func genesisBlock() *core.Block {
	header := &core.Header{
		Version: 1,
		DataHash: types.Hash{},
		Height: 0,
		Timestamp: 000000,
	}
	b, _ := core.NewBlock(header, nil)
	return b
}
