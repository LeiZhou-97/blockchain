package core

import (
	"testing"
	"time"

	"github.com/LeiZhou-97/blockchain/crypto"
	"github.com/LeiZhou-97/blockchain/types"
	"github.com/stretchr/testify/assert"
)

func TestSignBlock(t *testing.T) {
	privkey := crypto.GeneratePrivateKey()
	b := randomBlock(t, 0, types.Hash{})

	assert.Nil(t, b.Sign(privkey))
	assert.NotNil(t, b.Signature)

}

func TestVerifyBlock(t *testing.T) {
	privkey := crypto.GeneratePrivateKey()
	b := randomBlock(t, 0, types.Hash{})

	assert.Nil(t, b.Sign(privkey))
	assert.Nil(t, b.Verify())

	//TODO
	// edit header
	// b.Height = 100
	// assert.NotNil(t, b.Verify())
}

func randomBlock(t *testing.T, height uint32, prevBlockHash types.Hash) *Block {
	privKey := crypto.GeneratePrivateKey()
	tx := randomTxWithSignature(t)
	header := &Header{
		Version:       1,
		PrevBlockHash: prevBlockHash,
		Height:        height,
		Timestamp:     time.Now().UnixNano(),
	}

	b,err := NewBlock(header, []Transaction{tx})
	assert.Nil(t, err)
	dataHash, err := CalculateDataHash(b.Transactions)
	assert.Nil(t, err)
	b.Header.DataHash = dataHash
	assert.Nil(t, b.Sign(privKey))
	return b
}
