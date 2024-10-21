package core

import (
	"testing"
	"time"

	"github.com/LeiZhou-97/blockchain/crypto"
	"github.com/LeiZhou-97/blockchain/types"
	"github.com/stretchr/testify/assert"
)


func randomBlock(height uint32) *Block {
	header := &Header{
		Version: 1,
		PrevBlockHash: types.RandomHash(),
		Height: height,
		Timestamp: time.Now().UnixNano(),
	}

	tx := Transaction{
		Data: []byte("foo"),
	}
	return NewBlock(header, []Transaction{tx})
}

func TestSignBlock(t *testing.T) {
	privkey := crypto.GeneratePrivateKey()
	b := randomBlock(0)

	assert.Nil(t, b.Sign(privkey))
	assert.NotNil(t, b.Signature)

}

func TestVerifyBlock(t *testing.T) {
	privkey := crypto.GeneratePrivateKey()
	b := randomBlock(0)

	assert.Nil(t, b.Sign(privkey))
	assert.Nil(t, b.Verify())


	//TODO
	// edit header
	// b.Height = 100
	// assert.NotNil(t, b.Verify())
}
