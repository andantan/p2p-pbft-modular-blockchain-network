package core

import (
	"github.com/andantan/p2p-pbft-modular-blockchain-network/core/block"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestMemPool_PutAndCount(t *testing.T) {
	c := 10
	mp := GenerateTestMempool(t, c)

	for i := 0; i < c; i++ {
		tx := block.GenerateRandomTestTransaction(t)
		assert.NoError(t, mp.Put(tx))
	}

	assert.Equal(t, c, mp.Count())

	txFull := block.GenerateRandomTestTransaction(t)
	err := mp.Put(txFull)
	assert.ErrorIs(t, err, ErrMempoolFull)

	assert.Equal(t, c, mp.Count())
}

func TestMemPool_Contains(t *testing.T) {
	mp := GenerateTestMempool(t, 10)
	tx := block.GenerateRandomTestTransaction(t)

	assert.False(t, mp.Contains(tx))
	assert.NoError(t, mp.Put(tx))
	assert.True(t, mp.Contains(tx))
}

func TestMemPool_PruneAndClear(t *testing.T) {
	mp := GenerateTestMempool(t, 10)
	txx := make([]*block.Transaction, 5)
	for i := 0; i < 5; i++ {
		txx[i] = block.GenerateRandomTestTransaction(t)
		assert.NoError(t, mp.Put(txx[i]))
	}
	assert.Equal(t, 5, mp.Count())

	mp.Prune(txx[:2])
	assert.Equal(t, 3, mp.Count())
	assert.False(t, mp.Contains(txx[0]))
	assert.False(t, mp.Contains(txx[1]))
	assert.True(t, mp.Contains(txx[2]))
	assert.True(t, mp.Contains(txx[3]))
	assert.True(t, mp.Contains(txx[4]))

	mp.Clear()
	pendingTxs, err := mp.Pending()
	assert.ErrorIs(t, err, ErrMempoolEmpty)
	assert.Equal(t, 0, mp.Count())
	assert.Nil(t, pendingTxs)
}

func TestMemPool_Pending(t *testing.T) {
	mp := GenerateTestMempool(t, 10)
	tx1 := block.GenerateRandomTestTransaction(t)
	tx2 := block.GenerateRandomTestTransaction(t)

	assert.NoError(t, mp.Put(tx1))
	assert.NoError(t, mp.Put(tx2))

	pendingTxs, err := mp.Pending()
	assert.NoError(t, err)
	assert.NotNil(t, pendingTxs)
	assert.Equal(t, 2, len(pendingTxs))

	hash1, _ := tx1.Hash()
	hash2, _ := tx2.Hash()
	pendingHash1, _ := pendingTxs[0].Hash()
	pendingHash2, _ := pendingTxs[1].Hash()
	assert.True(t, hash1.Equal(pendingHash1))
	assert.True(t, hash2.Equal(pendingHash2))
}

func TestMemPool_RaceCondition(t *testing.T) {
	capacity := 100
	mp := GenerateTestMempool(t, capacity)
	var wg sync.WaitGroup
	numGoroutines := 20

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				tx := block.GenerateRandomTestTransaction(t)
				_ = mp.Put(tx) // ignore ErrMempoolFull
			}
		}()
	}

	wg.Wait()

	assert.LessOrEqual(t, mp.Count(), capacity)
}
