package core

import (
	"context"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/core/block"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestNewBlockchain(t *testing.T) {
	bc := GenerateTestBlockchain(t)
	assert.NotNil(t, bc.blockStorer)
	assert.NotNil(t, bc.blockProcessor)
	assert.Equal(t, uint64(0), bc.CurrentHeight())
	assert.True(t, bc.HasBlockHeight(0))
}

func TestBlockchain_AddBlock(t *testing.T) {
	bc := GenerateTestBlockchain(t)

	h0, err := bc.CurrentHeader()
	assert.NoError(t, err)

	b1 := block.GenerateRandomTestBlockWithPrevHeader(t, h0, 1<<5)
	h1, err := b1.Hash()
	assert.NoError(t, err)
	assert.NoError(t, bc.AddBlock(b1))
	assert.Equal(t, uint64(1), bc.CurrentHeight())
	assert.True(t, bc.HasBlockHash(h1))
	assert.True(t, bc.HasBlockHeight(bc.CurrentHeight()))

	// Case: ErrBlockKnown
	assert.ErrorIs(t, bc.AddBlock(b1), ErrBlockKnown)

	// Case: ErrFutureBlock
	b3 := block.GenerateRandomTestBlockWithHeight(t, 1<<3, 3)
	err = bc.AddBlock(b3)
	assert.ErrorIs(t, err, ErrFutureBlock)

	// Case: ErrUnknownParent
	b2Forked := block.GenerateRandomTestBlockWithHeight(t, 1<<5, 2)
	err = bc.AddBlock(b2Forked)
	assert.ErrorIs(t, err, ErrUnknownParent)
}

func TestBlockchain_Rollback(t *testing.T) {
	bc := GenerateTestBlockchain(t)

	AddTestBlocksToBlockchain(t, bc, 10)

	// Rollback to 5
	assert.NoError(t, bc.Rollback(5))
	assert.Equal(t, uint64(4), bc.CurrentHeight())
	assert.True(t, bc.HasBlockHeight(4))
	assert.False(t, bc.HasBlockHeight(5))

	AddTestBlocksToBlockchain(t, bc, 1)
	assert.Equal(t, uint64(5), bc.CurrentHeight())
}

func TestBlockchain_Clear(t *testing.T) {
	bc := GenerateTestBlockchain(t)

	AddTestBlocksToBlockchain(t, bc, 20)

	assert.NoError(t, bc.Clear())

	assert.Equal(t, uint64(0), bc.CurrentHeight())
	assert.False(t, bc.HasBlockHeight(5))
	assert.False(t, bc.HasBlockHeight(20))

	assert.Equal(t, uint64(0), bc.CurrentHeight())
	assert.True(t, bc.HasBlockHeight(0))
}

func TestBlockchain_ConcurrentReadAndWrite(t *testing.T) {
	bc := GenerateTestBlockchain(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var wg sync.WaitGroup

	// Write goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done(): // 테스트 종료 신호
				return
			case <-ticker.C:
				AddTestBlocksToBlockchain(t, bc, 1)
			}
		}
	}()

	// Read goroutine(5)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					h := bc.CurrentHeight()
					if h > 0 {
						_, _ = bc.GetBlockByHeight(h)
					}
					_ = bc.HasBlockHeight(h)
				}
			}
		}()
	}

	wg.Wait()
}
