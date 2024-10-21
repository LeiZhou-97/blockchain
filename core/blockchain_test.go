package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)


func newBlockChainWithGenesis(t *testing.T) *BlockChain {
	bc, err := NewBlockChain(randomBlock(0))
	assert.Nil(t, err)
	return bc
}

func TestAddBlock(t *testing.T) {
	bc := newBlockChainWithGenesis(t)

	for i:=0; i<1000; i++ {
		block := randomBlockWithSignature(t, uint32(i+1))
		assert.Nil(t, bc.AddBlock(block))
	}
	// height:1000
	// len:1001
	assert.Equal(t, bc.Height(), uint32(1000))

	// check validator
	assert.NotNil(t, bc.AddBlock(randomBlockWithSignature(t, uint32(88))))
}

func TestBlockChain(t *testing.T) {
	bc := newBlockChainWithGenesis(t)
	assert.NotNil(t, bc.validator)
	assert.Equal(t, bc.Height(), uint32(0))	
}

func TestHashBlock(t *testing.T) {
	bc := newBlockChainWithGenesis(t)

	assert.True(t, bc.HasBlock(0))
}