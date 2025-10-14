package core

import (
	"errors"
	"github.com/andantan/modular-blockchain/core/block"
	"github.com/andantan/modular-blockchain/types"
	"github.com/andantan/modular-blockchain/util"
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
	if !b.IsConsented() {

	}

	height := b.Header.Height

	if height == 0 {
		return b.Verify()
	}

	if p.chain.HasBlockHeight(height) {
		return ErrBlockKnown
	}

	if p.chain.GetCurrentHeight()+1 != height {
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

	if !prevBlockHash.Equal(b.Header.PrevBlockHash) {
		return ErrUnknownParent
	}

	if err = b.Verify(); err != nil {
		return err
	}

	return nil
}
