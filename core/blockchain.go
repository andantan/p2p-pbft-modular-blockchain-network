package core

import (
	"github.com/andantan/p2p-pbft-modular-blockchain-network/core/block"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/core/contract"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/types"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/util"
	"github.com/go-kit/log"
)

type Blockchain struct {
	logger log.Logger

	blockStorer    Storer
	blockProcessor Processor
	contract       contract.Contract
}

func NewBlockchain() *Blockchain {
	return &Blockchain{
		logger: util.LoggerWithPrefixes("Blockchain"),
	}
}

func (bc *Blockchain) SetStorer(s Storer) {
	bc.blockStorer = s
}

func (bc *Blockchain) SetBlockProcessor(p Processor) {
	bc.blockProcessor = p
}

func (bc *Blockchain) SetContract(c contract.Contract) {
	bc.contract = c
}

func (bc *Blockchain) Bootstrap() {
	if bc.blockStorer == nil {
		panic("Storer is nil")
	}

	if bc.blockProcessor == nil {
		panic("Processor is nil")
	}

	//if bc.contract == nil {
	//	panic("Contract is nil")
	//}

	if err := bc.blockStorer.LoadStorage(); err != nil {
		panic(err)
	}

	height := bc.CurrentHeight()

	if height == 0 {
		if err := bc.AddBlock(block.NewGenesisBlock()); err != nil {
			panic(err)
		}

		return
	}

	for h := uint64(0); h <= height; h++ {
		b, err := bc.GetBlockByHeight(h)

		if err != nil {
			panic(err)
		}

		if err = bc.blockProcessor.ProcessBlock(b); err != nil {
			panic(err)
		}

		hash, _ := b.Hash()
		_ = bc.logger.Log("msg", "verified block", "height", height, "hash", hash.String())
	}
}

func (bc *Blockchain) AddBlock(b *block.Block) error {
	var (
		err  error
		hash types.Hash
	)

	if err = bc.blockProcessor.ProcessBlock(b); err != nil {
		return err
	}

	if hash, err = b.Hash(); err != nil {
		return err
	}

	_ = bc.logger.Log(
		"msg", "new block added",
		"hash", hash.ShortString(8),
		"height", b.GetHeight(),
		"weight", b.GetWeight(),
	)

	return bc.blockStorer.StoreBlock(b)
}

func (bc *Blockchain) Rollback(from uint64) error {
	ch := bc.CurrentHeight()

	if ch < from {
		return nil
	}

	_ = bc.logger.Log("msg", "rolling back chain", "from", from, "to", ch)

	for h := ch; h >= from; h-- {
		if err := bc.blockStorer.RemoveBlock(h); err != nil {
			_ = bc.logger.Log(
				"error", "failed to remove block from storage",
				"height", h,
				"err", err,
			)
			return err
		}
	}

	return nil
}

func (bc *Blockchain) Clear() error {
	if err := bc.blockStorer.ClearStorage(); err != nil {
		return err
	}

	if err := bc.AddBlock(block.NewGenesisBlock()); err != nil {
		return err
	}

	return nil
}

func (bc *Blockchain) CurrentHeight() uint64 {
	return bc.blockStorer.CurrentHeight()
}

func (bc *Blockchain) CurrentBlock() (*block.Block, error) {
	return bc.GetBlockByHeight(bc.CurrentHeight())
}

func (bc *Blockchain) CurrentHeader() (*block.Header, error) {
	return bc.GetHeaderByHeight(bc.CurrentHeight())
}

func (bc *Blockchain) GetBlockByHeight(h uint64) (*block.Block, error) {
	return bc.blockStorer.GetBlockByHeight(h)
}

func (bc *Blockchain) GetBlockByHash(h types.Hash) (*block.Block, error) {
	return bc.blockStorer.GetBlockByHash(h)
}

func (bc *Blockchain) GetHeaderByHeight(h uint64) (*block.Header, error) {
	return bc.blockStorer.GetHeaderByHeight(h)
}

func (bc *Blockchain) GetHeaderByHash(h types.Hash) (*block.Header, error) {
	return bc.blockStorer.GetHeaderByHash(h)
}

func (bc *Blockchain) HasBlockHash(h types.Hash) bool {
	return bc.blockStorer.HasBlockHash(h)
}

func (bc *Blockchain) HasBlockHeight(h uint64) bool {
	return bc.blockStorer.HasBlockHeight(h)
}
