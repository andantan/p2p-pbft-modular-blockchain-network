package core

import (
	"github.com/andantan/p2p-pbft-modular-blockchain-network/core/block"
	"github.com/stretchr/testify/assert"
	"testing"
)

func GenerateTestBlockchain(t *testing.T) *Blockchain {
	t.Helper()

	dir := t.TempDir()
	s := NewBlockStorage(dir)

	bc := NewBlockchain()
	bc.SetStorer(s)

	p := NewBlockProcessor(bc)
	bc.SetBlockProcessor(p)

	bc.Bootstrap()

	return bc
}

func GenerateTestBlockchainAndProcessor(t *testing.T) (*Blockchain, *BlockProcessor) {
	t.Helper()

	dir := t.TempDir()
	s := NewBlockStorage(dir)

	bc := NewBlockchain()
	bc.SetStorer(s)

	p := NewBlockProcessor(bc)
	bc.SetBlockProcessor(p)

	bc.Bootstrap()

	return bc, p
}

func AddTestBlocksToBlockchain(t *testing.T, bc *Blockchain, c uint64) {
	t.Helper()

	beforeHeight := bc.GetCurrentHeight()

	for i := uint64(1); i < c+1; i++ {
		hr, err := bc.GetCurrentHeader()
		assert.NoError(t, err)

		b := block.GenerateRandomTestBlockWithPrevHeader(t, hr, 1<<3)
		h, err := b.Hash()
		assert.NoError(t, err)
		assert.NoError(t, bc.AddBlock(b))
		assert.Equal(t, i+beforeHeight, bc.GetCurrentHeight())
		assert.True(t, bc.HasBlockHash(h))
		assert.True(t, bc.HasBlockHeight(bc.GetCurrentHeight()))
	}
}

func GenerateTestMempool(t *testing.T, c int) *MemPool {
	t.Helper()

	m := NewMemPool(c)
	assert.Equal(t, 0, m.Count())

	return m
}
