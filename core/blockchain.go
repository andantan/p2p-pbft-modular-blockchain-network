package core

import (
	"github.com/andantan/modular-blockchain/core/block"
	"github.com/andantan/modular-blockchain/core/contract"
	"github.com/andantan/modular-blockchain/types"
	"github.com/andantan/modular-blockchain/util"
	"github.com/go-kit/log"
)

type Chain interface {
	Bootstrap()
	AddBlock(*block.Block) error
	Rollback(uint64) error
	Clear() error

	GetCurrentHeight() uint64
	GetCurrentBlock() (*block.Block, error)
	GetCurrentHeader() (*block.Header, error)
	GetBlockByHeight(h uint64) (*block.Block, error)
	GetBlockByHash(h types.Hash) (*block.Block, error)
	GetHeaderByHeight(h uint64) (*block.Header, error)
	GetHeaderByHash(h types.Hash) (*block.Header, error)
	HasBlockHash(h types.Hash) bool
	HasBlockHeight(h uint64) bool
}

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
	_ = bc.logger.Log("msg", "start blockchain bootstrap")

	if bc.blockStorer == nil || bc.blockProcessor == nil {
		panic("storer or processor is not initialized")
	}

	//if bc.contract == nil {
	//	panic("Contract is nil")
	//}

	if err := bc.blockStorer.LoadStorage(); err != nil {
		panic(err)
	}

	height := bc.GetCurrentHeight()
	_ = bc.logger.Log("msg", "storage loaded", "height", height)

	if height == 0 {
		_ = bc.logger.Log("msg", "chain is empty, creating genesis block")

		if err := bc.AddBlock(block.NewGenesisBlock()); err != nil {
			panic(err)
		}

		_ = bc.logger.Log("msg", "finished blockchain bootstrap", "final_height", bc.GetCurrentHeight())

		return
	}

	_ = bc.logger.Log("msg", "verifying existing blocks", "count", height+1)

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

	_ = bc.logger.Log("msg", "finished blockchain bootstrap", "final_height", bc.GetCurrentHeight())

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
		"height", b.Header.Height,
		"weight", b.Header.Weight,
	)

	return bc.blockStorer.StoreBlock(b)
}

func (bc *Blockchain) Rollback(from uint64) error {
	ch := bc.blockStorer.CurrentHeight()

	if ch < from {
		return nil
	}

	_ = bc.logger.Log("msg", "rolling back chain", "from", from, "to", ch)

	for h := ch; h >= from; h-- {
		if err := bc.blockStorer.RemoveBlock(h); err != nil {
			_ = bc.logger.Log("msg", "failed to remove block from storage", "height", h, "err", err)
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

func (bc *Blockchain) GetCurrentHeight() uint64 {
	return bc.blockStorer.CurrentHeight()
}

func (bc *Blockchain) GetCurrentBlock() (*block.Block, error) {
	return bc.blockStorer.CurrentBlock()
}

func (bc *Blockchain) GetCurrentHeader() (*block.Header, error) {
	return bc.blockStorer.CurrentHeader()
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
