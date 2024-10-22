package network

import (
	"fmt"
	"time"

	"github.com/LeiZhou-97/blockchain/core"
	"github.com/LeiZhou-97/blockchain/crypto"
	"github.com/sirupsen/logrus"
)

type ServerOpts struct {
	Transports []Transport
	BlockTIme  time.Duration
	PrivateKey *crypto.PrivateKey
}

type Server struct {
	ServerOpts
	blockTime   time.Duration
	memPool     *TxPool
	isValidator bool
	rpcCh       chan RPC
	quitCh      chan struct{}
}

func NewServer(opts ServerOpts) *Server {
	return &Server{
		ServerOpts:  opts,
		blockTime:   opts.BlockTIme,
		memPool:     NewTxPool(),
		isValidator: opts.PrivateKey != nil,
		rpcCh:       make(chan RPC),
		quitCh:      make(chan struct{}, 1),
	}
}

func (s *Server) Start() {
	s.initTransport()
	ticker := time.NewTicker(5 * time.Second)

free:
	for {
		select {
		case rpc := <-s.rpcCh:
			fmt.Printf("%+v", rpc)
		case <-s.quitCh:
			break free
		case <-ticker.C:
			if s.isValidator {
				s.createNewBlock()
			}
		}
	}

	fmt.Println("Server shutdown")
}

func (s *Server) handleTransaction(tx *core.Transaction) error {
	if err := tx.Verify(); err != nil {
		return err
	}

	hash := tx.Hash(core.TxHasher{})
	if s.memPool.Has(hash) {
		logrus.WithFields(logrus.Fields{
			"hash": hash,
		}).Info("transaction already in mempool")
	}
	logrus.WithFields(logrus.Fields{
		"hash": hash,
	}).Info("adding new tx to the mempool")


	return s.memPool.Add(tx)
}

func (s *Server) createNewBlock() error {
	fmt.Println("creating a new block")
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
