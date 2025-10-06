package core

import (
	"errors"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/core/block"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/types"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/util"
	"github.com/go-kit/log"
)

var (
	ErrBlockKnown    = errors.New("block already known")
	ErrFutureBlock   = errors.New("block is too high")
	ErrUnknownParent = errors.New("chain has been forked")
)

type Processor interface {
	ProcessBlock(*block.Block) error
}

type BlockProcessor struct {
	logger log.Logger
	chain  *Blockchain
}

func NewBlockProcessor(chain *Blockchain) *BlockProcessor {
	return &BlockProcessor{
		logger: util.LoggerWithPrefixes("Processor"),
		chain:  chain,
	}
}

func (p *BlockProcessor) ProcessBlock(b *block.Block) error {
	height := b.GetHeight()

	if height == 0 {
		return b.Verify()
	}

	if p.chain.HasBlockHeight(height) {
		return ErrBlockKnown
	}

	if p.chain.CurrentHeight()+1 != height {
		return ErrFutureBlock
	}

	var (
		err           error
		prevHeader    *block.Header
		prevBlockHash types.Hash
	)

	if prevHeader, err = p.chain.GetHeaderByHeight(height - 1); err != nil {
		return err
	}

	if prevBlockHash, err = prevHeader.Hash(); err != nil {
		return err
	}

	if !prevBlockHash.Eq(b.Header.PrevBlockHash) {
		return ErrUnknownParent
	}

	if err = b.Verify(); err != nil {
		return err
	}

	return nil
}
