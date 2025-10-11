package core

import (
	"encoding/hex"
	"fmt"
	"github.com/andantan/modular-blockchain/codec"
	"github.com/andantan/modular-blockchain/core/block"
	"github.com/andantan/modular-blockchain/types"
	"github.com/andantan/modular-blockchain/util"
	"github.com/go-kit/log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Storer interface {
	CurrentHeight() uint64
	CurrentBlock() (*block.Block, error)
	CurrentHeader() (*block.Header, error)

	LoadStorage() error
	ClearStorage() error

	StoreBlock(*block.Block) error
	RemoveBlock(uint64) error

	GetBlockByHash(types.Hash) (*block.Block, error)
	GetBlockByHeight(uint64) (*block.Block, error)
	GetHeaderByHash(types.Hash) (*block.Header, error)
	GetHeaderByHeight(uint64) (*block.Header, error)

	HasBlockHash(types.Hash) bool
	HasBlockHeight(uint64) bool
}

type BlockStorage struct {
	logger log.Logger
	dir    string
	ext    string
	ioLock sync.RWMutex
	set    *types.AtomicSet[types.Hash]
	index  *types.AtomicMap[uint64, types.Hash]
}

func NewBlockStorage(dir string) *BlockStorage {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		panic(err)
	}

	return &BlockStorage{
		logger: util.LoggerWithPrefixes("Storage"),
		dir:    dir,
		ext:    ".block",
		set:    types.NewAtomicSet[types.Hash](),
		index:  types.NewAtomicMap[uint64, types.Hash](),
	}
}

func (bs *BlockStorage) CurrentHeight() uint64 {
	c := uint64(bs.index.Len())

	if c == 0 {
		return 0
	}

	return c - 1
}

func (bs *BlockStorage) CurrentBlock() (*block.Block, error) {
	return bs.GetBlockByHeight(bs.CurrentHeight())
}

func (bs *BlockStorage) CurrentHeader() (*block.Header, error) {
	return bs.GetHeaderByHeight(bs.CurrentHeight())
}

func (bs *BlockStorage) LoadStorage() error {
	bs.ioLock.RLock()
	defer bs.ioLock.RUnlock()

	files, err := os.ReadDir(bs.dir)

	if err != nil {
		panic(err)
	}

	_ = bs.logger.Log("msg", "loading blocks from storage...")

	for _, file := range files {
		var (
			blockBytes []byte
			b          = new(block.Block)
			h          types.Hash
		)

		if file.IsDir() || !strings.HasSuffix(file.Name(), bs.ext) {
			continue
		}

		hashStr := strings.TrimSuffix(file.Name(), bs.ext)

		if len(hashStr) != types.HashLength*2 {
			_ = bs.logger.Log("msg", "found invalid block file name", "file", file.Name())
			continue
		}

		filePath := filepath.Join(bs.dir, file.Name())

		if blockBytes, err = os.ReadFile(filePath); err != nil {
			_ = bs.logger.Log("msg", "failed to read block file", "path", filePath, "err", err)
			continue
		}

		if err = codec.DecodeProto(blockBytes, b); err != nil {
			_ = bs.logger.Log("msg", "failed to decode block file", "path", filePath, "err", err)
			continue
		}

		if h, err = b.Hash(); err != nil {
			return err
		}

		bs.set.Put(h)
		bs.index.Put(b.Header.Height, h)
	}

	return nil
}

func (bs *BlockStorage) ClearStorage() error {
	bs.ioLock.Lock()
	defer bs.ioLock.Unlock()

	if err := os.RemoveAll(bs.dir); err != nil {
		return err
	}

	bs.set = types.NewAtomicSet[types.Hash]()
	bs.index = types.NewAtomicMap[uint64, types.Hash]()

	return os.MkdirAll(bs.dir, os.ModePerm)
}

func (bs *BlockStorage) StoreBlock(b *block.Block) error {
	var (
		err        error
		blockBytes []byte
		blockHash  types.Hash
	)

	if blockBytes, err = codec.EncodeProto(b); err != nil {
		return err
	}

	if blockHash, err = b.Hash(); err != nil {
		return err
	}

	n := fmt.Sprintf("%s%s", hex.EncodeToString(blockHash.Bytes()), bs.ext)
	p := filepath.Join(bs.dir, n)

	bs.ioLock.Lock()
	defer bs.ioLock.Unlock()

	err = os.WriteFile(p, blockBytes, os.ModePerm)

	if err != nil {
		return err
	}

	h, _ := b.Hash()

	bs.set.Put(h)
	bs.index.Put(b.Header.Height, h)

	return nil
}

func (bs *BlockStorage) RemoveBlock(height uint64) error {
	bs.ioLock.Lock()
	defer bs.ioLock.Unlock()

	h, ok := bs.index.Get(height)

	if !ok {
		return nil
	}

	n := fmt.Sprintf("%s%s", hex.EncodeToString(h.Bytes()), bs.ext)
	p := filepath.Join(bs.dir, n)

	if err := os.Remove(p); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}

	bs.set.Remove(h)
	bs.index.Remove(height)

	return nil
}

func (bs *BlockStorage) GetBlockByHash(h types.Hash) (*block.Block, error) {
	n := fmt.Sprintf("%s%s", hex.EncodeToString(h.Bytes()), bs.ext)
	p := filepath.Join(bs.dir, n)

	bs.ioLock.RLock()
	defer bs.ioLock.RUnlock()

	data, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}

	b := new(block.Block)
	if err = codec.DecodeProto(data, b); err != nil {
		return nil, err
	}

	return b, nil
}

func (bs *BlockStorage) GetBlockByHeight(height uint64) (*block.Block, error) {
	h, ok := bs.index.Get(height)
	if !ok {
		return nil, fmt.Errorf("block with height %d not found", height)
	}

	return bs.GetBlockByHash(h)
}

func (bs *BlockStorage) GetHeaderByHash(h types.Hash) (*block.Header, error) {
	b, err := bs.GetBlockByHash(h)

	if err != nil {
		return nil, err
	}

	return b.Header, nil
}

func (bs *BlockStorage) GetHeaderByHeight(height uint64) (*block.Header, error) {
	b, err := bs.GetBlockByHeight(height)

	if err != nil {
		return nil, err
	}

	return b.Header, nil
}

func (bs *BlockStorage) HasBlockHash(h types.Hash) bool {
	return bs.set.Contains(h)
}

func (bs *BlockStorage) HasBlockHeight(height uint64) bool {
	return bs.index.Exists(height)
}
