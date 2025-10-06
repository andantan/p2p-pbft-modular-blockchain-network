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

func NewBlockchain(bs Storer, bp Processor) *Blockchain {
	if bs == nil || bp == nil {
		panic("storer & processor is nil")
	}

	bc := &Blockchain{
		logger:         util.LoggerWithPrefixes("Blockchain"),
		blockStorer:    bs,
		blockProcessor: bp,
		contract:       nil,
	}

	bc.bootstrap()

	return bc
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

func (bc *Blockchain) CurrentHeight() uint64 {
	return bc.blockStorer.CurrentHeight()
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

func (bc *Blockchain) Clear() error {
	return bc.blockStorer.ClearStorage()
}

func (bc *Blockchain) bootstrap() {
	if err := bc.blockStorer.LoadStorage(); err != nil {
		panic(err)
	}

	panic("todo!")
}
