package core

import "fmt"

type Validator interface {
	ValidateBlock(*Block) error
}

type BlockValidator struct {
	bc *BlockChain
}

func NewBlockValidator(bc *BlockChain) *BlockValidator {
	return &BlockValidator{
		bc: bc,
	}
}

func (v *BlockValidator) ValidateBlock(b *Block) error {
	if v.bc.HasBlock(b.Height) {
		return fmt.Errorf("chain already contains blcok (%d) with hash (%s)", b.Height, b.Hash(BlockHasher{}))
	}

	if b.Height != v.bc.Height()+1 {
		return fmt.Errorf("block (%s) with height (%d) is too high ==> current height (%d)", b.Hash(BlockHasher{}), b.Height, v.bc.Height())
	}

	prevHeader, err := v.bc.GetHeader(b.Height - 1)
	if err != nil {
		return err
	}

	hash := BlockHasher{}.Hash(prevHeader)
	if hash != b.PrevBlockHash {
		return fmt.Errorf("the hash of the previous block is invalid")
	}

	if err := b.Verify(); err != nil {
		return err
	}
	return nil
}
