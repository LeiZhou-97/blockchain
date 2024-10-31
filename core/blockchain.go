package core

import (
	"fmt"
	"sync"

	"github.com/LeiZhou-97/blockchain/types"
	"github.com/go-kit/log"
)

type BlockChain struct {
	logger    log.Logger
	store     Storage
	lock      sync.RWMutex
	headers   []*Header
	blocks    []*Block
	txStore map[types.Hash]*Transaction
	blockStore map[types.Hash]*Block
	validator Validator
	// TODO make this an interface
	contractState *State
}

func NewBlockChain(l log.Logger, genesis *Block) (*BlockChain, error) {
	bc := &BlockChain{
		headers: []*Header{},
		store:   NewMemStore(),
		logger: l,
		contractState: NewState(),
		blockStore: make(map[types.Hash]*Block),
		txStore: make(map[types.Hash]*Transaction),
	}

	bc.validator = NewBlockValidator(bc)

	err := bc.addBlockWithoutValidation(genesis)

	return bc, err
}

func (bc *BlockChain) SGetValidator(v Validator) {
	bc.validator = v
}

func (bc *BlockChain) AddBlock(b *Block) error {
	// validate before adding to chain
	if err := bc.validator.ValidateBlock(b); err != nil {
		return err
	}
	for _, tx := range b.Transactions {
		bc.logger.Log("msg", "executing code", "hash", tx.Hash(&TxHasher{}))
		vm := NewVM(tx.Data, bc.contractState)
		if err := vm.Run(); err != nil {
			return err
		}
		
		result := vm.stack.Pop()

		bc.logger.Log("vm result", result)
	}
	return bc.addBlockWithoutValidation(b)
}

func (bc *BlockChain) GetHeader(height uint32) (*Header, error) {
	if height > bc.Height() {
		return nil, fmt.Errorf("height (%d) too high", height)
	}

	bc.lock.Lock()
	defer bc.lock.Unlock()
	return bc.headers[int(height)], nil
}


func (bc *BlockChain) GetBlockByHash(hash types.Hash) (*Block, error) {
	block, ok := bc.blockStore[hash]
	if !ok {
		return nil, fmt.Errorf("block with hash (%s) not exist", hash)
	}
	return block, nil
}

func (bc *BlockChain) GetTxByHash(hash types.Hash) (*Transaction, error) {
	bc.lock.Lock()
	defer bc.lock.Unlock()
	tx, ok := bc.txStore[hash]
	if !ok {
		return nil, fmt.Errorf("could not find tx with hash (%s)", hash)
	}
	return tx, nil
}

func (bc *BlockChain) GetBlock(height uint32) (*Block, error) {
	if height > bc.Height() {
		return nil, fmt.Errorf("height (%d) too high", height)
	}

	bc.lock.Lock()
	defer bc.lock.Unlock()
	return bc.blocks[int(height)], nil
}

func (bc *BlockChain) HasBlock(height uint32) bool {
	return height <= bc.Height()
}

func (bc *BlockChain) Height() uint32 {
	bc.lock.RLock()
	defer bc.lock.RUnlock()

	return uint32(len(bc.headers) - 1)
}

func (bc *BlockChain) addBlockWithoutValidation(b *Block) error {
	bc.lock.Lock()
	bc.headers = append(bc.headers, b.Header)
	bc.blocks = append(bc.blocks, b)
	bc.blockStore[b.Hash(BlockHasher{})] = b
	bc.lock.Unlock()

	for _, tx := range b.Transactions {
		bc.txStore[tx.Hash(TxHasher{})] = tx
	}

	bc.logger.Log(
		"msg", "adding new block",
		"hash", b.Hash(BlockHasher{}),
		"height", b.Height,
		"transactions", len(b.Transactions),
	)

	return bc.store.Put(b)
}
