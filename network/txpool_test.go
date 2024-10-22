package network

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/LeiZhou-97/blockchain/core"
	"github.com/stretchr/testify/assert"
)

func TestTxPool(t *testing.T) {
	p := NewTxPool()
	assert.Equal(t, p.Len(), 0)
}

func TestTxPoolAddTx(t *testing.T) {
	p := NewTxPool()
	tx := core.NewTransaction([]byte("foo"))
	assert.Nil(t, p.Add(tx))
	assert.Equal(t, p.Len(), 1)

	txx := core.NewTransaction([]byte("foo"))
	_ = p.Add(txx)
	assert.Equal(t, p.Len(), 1)

	p.Flush()
	assert.Equal(t, p.Len(), 0)
}

func TestSortTransactions (t *testing.T) {
	p := NewTxPool()

	txLen := 5

	for i:=0; i<txLen; i++ {
		tx := core.NewTransaction([]byte(strconv.FormatInt(int64(i), 10)))
		tx.SetFirstSeen(int64(i*rand.Intn(1000)))
		assert.Nil(t, p.Add(tx))
	}
	assert.Equal(t, txLen, p.Len())

	txx := p.Transactions()
	for i:=0; i<txLen-1; i++ {
		assert.True(t, txx[i].FirstSeen() < txx[i+1].FirstSeen())
	}
}