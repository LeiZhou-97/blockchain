package core

import (
	"crypto/sha256"

	"github.com/LeiZhou-97/blockchain/types"
)



type Hasher[T any] interface {
	Hash(T) types.Hash
}

type BlockHasher struct {

}

func (bh BlockHasher) Hash(b *Header) types.Hash {
	h := sha256.Sum256(b.Bytes())
	return types.Hash(h)
}


type TxHasher struct {}

func (TxHasher) Hash(tx *Transaction) types.Hash {
	return types.Hash(sha256.Sum256(tx.Data))
}