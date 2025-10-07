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

func AddTestBlocks(t *testing.T, bc *Blockchain, amount uint64) {
	t.Helper()

	beforeHeight := bc.CurrentHeight()

	for i := uint64(1); i < amount+1; i++ {
		hr, err := bc.CurrentHeader()
		assert.NoError(t, err)

		b := block.GenerateRandomTestBlockWithPrevHeader(t, hr, 1<<3)
		h, err := b.Hash()
		assert.NoError(t, err)
		assert.NoError(t, bc.AddBlock(b))
		assert.Equal(t, i+beforeHeight, bc.CurrentHeight())
		assert.True(t, bc.HasBlockHash(h))
		assert.True(t, bc.HasBlockHeight(bc.CurrentHeight()))
	}
}
