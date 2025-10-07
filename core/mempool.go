package core

import (
	"errors"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/core/block"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/types"
	"sync"
)

var (
	ErrMempoolFull = errors.New("mempool is full")
)

type VirtualMemoryPool interface {
	Put(*block.Transaction) error
	Contains(*block.Transaction) bool
	Prune([]*block.Transaction)
	Pending() []*block.Transaction
	Count() int
	Clear()
}

type MemPool struct {
	lock     sync.RWMutex
	pool     *types.AtomicCache[types.Hash, *block.Transaction]
	capacity int
}

func NewMemPool(capacity int) *MemPool {
	return &MemPool{
		pool:     types.NewAtomicCache[types.Hash, *block.Transaction](),
		capacity: capacity,
	}
}

func (mp *MemPool) Put(tx *block.Transaction) error {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	if mp.pool.Count() >= mp.capacity {
		return ErrMempoolFull
	}

	hash, err := tx.Hash()
	if err != nil {
		return err
	}

	mp.pool.Put(hash, tx)
	return nil
}

func (mp *MemPool) Contains(tx *block.Transaction) bool {
	mp.lock.RLock()
	defer mp.lock.RUnlock()

	h, _ := tx.Hash()
	return mp.pool.Contains(h)
}

func (mp *MemPool) Prune(txx []*block.Transaction) {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	r := make([]types.Hash, len(txx))

	for i, tx := range txx {
		h, _ := tx.Hash()
		r[i] = h
	}

	mp.pool.Prune(r)
}

func (mp *MemPool) Pending() []*block.Transaction {
	mp.lock.RLock()
	defer mp.lock.RUnlock()

	return mp.pool.Values()
}

func (mp *MemPool) Count() int {
	mp.lock.RLock()
	defer mp.lock.RUnlock()

	return mp.pool.Count()
}

func (mp *MemPool) Clear() {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	mp.pool.Clear()
}
