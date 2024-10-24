package core

import (
	"testing"

	"github.com/LeiZhou-97/blockchain/types"
	"github.com/stretchr/testify/assert"
)


func TestAddBlock(t *testing.T) {
	bc := newBlockChainWithGenesis(t)

	for i:=0; i<1000; i++ {
		block := randomBlock(t, uint32(i+1), getPrevBlockHash(t, bc, uint32(i+1)))
		assert.Nil(t, bc.AddBlock(block))
	}
	// height:1000
	// len:1001
	assert.Equal(t, bc.Height(), uint32(1000))

	// check validator
	assert.NotNil(t, bc.AddBlock(randomBlock(t, uint32(88), types.Hash{})))
}

func TestBlockChain(t *testing.T) {
	bc := newBlockChainWithGenesis(t)
	assert.NotNil(t, bc.validator)
	assert.Equal(t, bc.Height(), uint32(0))	
}

func TestHashBlock(t *testing.T) {
	bc := newBlockChainWithGenesis(t)

	assert.True(t, bc.HasBlock(0))
	assert.False(t, bc.HasBlock(1))
}

func TestGetHeader(t *testing.T) {
	bc := newBlockChainWithGenesis(t)
	lenBlocks := 1000

	for i:=0; i<lenBlocks; i++ {
		block := randomBlock(t, uint32(i+1), getPrevBlockHash(t, bc, uint32(i+1)))
		assert.Nil(t, bc.AddBlock(block))
		header, err := bc.GetHeader(block.Height)
		assert.Nil(t, err)
		assert.Equal(t, header, block.Header)
	}
}

func TestAddBlockToHigh(t *testing.T) {
	bc := newBlockChainWithGenesis(t)

	assert.Nil(t, bc.AddBlock(randomBlock(t, uint32(1), getPrevBlockHash(t, bc, uint32(1)))))
	assert.NotNil(t, bc.AddBlock(randomBlock(t, uint32(3), types.Hash{})))
}

func newBlockChainWithGenesis(t *testing.T) *BlockChain {
	bc, err := NewBlockChain(randomBlock(t, 0, types.Hash{}))
	assert.Nil(t, err)
	return bc
}

func getPrevBlockHash(t *testing.T, bc *BlockChain, height uint32) types.Hash {
	prevHeader, err := bc.GetHeader(height-1)
	assert.Nil(t, err)

	return BlockHasher{}.Hash(prevHeader)
}

