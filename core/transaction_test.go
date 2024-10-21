package core

import (
	"testing"

	"github.com/LeiZhou-97/blockchain/crypto"
	"github.com/stretchr/testify/assert"
)


func TestSignTransaction(t *testing.T) {
	data := []byte("foo")
	privKey := crypto.GeneratePrivateKey()
	tx := &Transaction{
		Data: data,
	}

	assert.Nil(t, tx.Sign(privKey))
	assert.NotNil(t, tx.Signature)
}


func TestVerifyTransaction(t *testing.T) {
	data := []byte("foo")
	privKey := crypto.GeneratePrivateKey()
	tx := &Transaction{
		Data: data,
	}

	assert.Nil(t, tx.Sign(privKey))
	assert.Nil(t, tx.Verify())
}
