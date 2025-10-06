package core

import (
	"github.com/andantan/p2p-pbft-modular-blockchain-network/core/block"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func newTestBlockStorage(t *testing.T) *BlockStorage {
	t.Helper()

	dir := t.TempDir()
	bs := NewBlockStorage(dir)
	return bs
}

func TestBlockStorage_StoreAndGetBlock(t *testing.T) {
	bs := newTestBlockStorage(t)
	b1 := block.GenerateRandomBlockWithHeight(t, 1<<10, 0)

	assert.NoError(t, bs.StoreBlock(b1))

	b1Read, err := bs.GetBlockByHeight(b1.GetHeight())
	assert.NoError(t, err)

	b1Hash, _ := b1.Hash()
	b1ReadHash, _ := b1Read.Hash()
	assert.True(t, b1Hash.Eq(b1ReadHash))

	b1ReadByHash, err := bs.GetBlockByHash(b1Hash)
	assert.NoError(t, err)
	b1ReadByHashHash, _ := b1ReadByHash.Hash()
	assert.True(t, b1Hash.Eq(b1ReadByHashHash))
}

func TestBlockStorer_LoadStorage(t *testing.T) {
	dir := t.TempDir()

	bs1 := NewBlockStorage(dir)
	b0 := block.GenerateRandomBlock(t, 1<<11)
	b1 := block.GenerateRandomBlockWithHeight(t, 1<<13, b0.GetHeight()+1)
	assert.NoError(t, bs1.StoreBlock(b0))
	assert.NoError(t, bs1.StoreBlock(b1))

	bs2 := NewBlockStorage(dir)
	assert.NoError(t, bs2.LoadStorage())

	assert.Equal(t, uint64(1), bs2.CurrentHeight())
	assert.True(t, bs2.HasBlockHeight(b0.GetHeight()))

	b0Hash, err := b0.Hash()
	assert.NoError(t, err)
	assert.True(t, bs2.HasBlockHash(b0Hash))

	b1Hash, err := b1.Hash()
	assert.NoError(t, err)
	assert.True(t, bs2.HasBlockHash(b1Hash))
}

func TestBlockStorer_ClearAndRemoveBlock(t *testing.T) {
	bs := newTestBlockStorage(t)
	b0 := block.GenerateRandomBlockWithHeight(t, 0, 0)
	b1 := block.GenerateRandomBlockWithHeight(t, 1<<15, 1)
	assert.NoError(t, bs.StoreBlock(b0))
	assert.NoError(t, bs.StoreBlock(b1))

	assert.NoError(t, bs.RemoveBlock(0))
	assert.False(t, bs.HasBlockHeight(0))
	assert.True(t, bs.HasBlockHeight(1))
	assert.Equal(t, 1, bs.index.Len())

	assert.NoError(t, bs.ClearStorage())
	assert.Equal(t, 0, bs.index.Len())
	assert.Equal(t, 0, bs.set.Len())
}

func TestBlockStorer_GetHeader(t *testing.T) {
	bs := newTestBlockStorage(t)
	b := block.GenerateRandomBlockWithHeight(t, 1<<4, 10)
	assert.NoError(t, bs.StoreBlock(b))
	blockHash, err := b.Hash()
	assert.NoError(t, err)

	// 1. GetHeaderByHeight
	h, err := bs.GetHeaderByHeight(10)
	assert.NoError(t, err)
	hHash1, err := h.Hash()
	assert.Equal(t, blockHash, hHash1)

	// 2. GetHeaderByHash
	bHash, _ := b.Hash()
	h, err = bs.GetHeaderByHash(bHash)
	assert.NoError(t, err)
	hHash2, err := h.Hash()
	assert.Equal(t, blockHash, hHash2)
}

func TestBlockStorer_HasBlock(t *testing.T) {
	bs := newTestBlockStorage(t)
	b := block.GenerateRandomBlockWithHeight(t, 1<<7, 5)
	assert.NoError(t, bs.StoreBlock(b))

	bHash, err := b.Hash()
	assert.NoError(t, err)

	assert.True(t, bs.HasBlockHeight(5))
	assert.True(t, bs.HasBlockHash(bHash))

	assert.False(t, bs.HasBlockHeight(99))
	assert.False(t, bs.HasBlockHash(util.RandomHash()))
}
