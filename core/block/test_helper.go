package block

import (
	"github.com/andantan/p2p-pbft-modular-blockchain-network/codec"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/util"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func GenerateRandomTestTransaction(t *testing.T) *Transaction {
	t.Helper()

	privKey, _ := crypto.GenerateTestKeyPair(t)

	tx := NewTransaction(util.RandomBytes(1<<12), util.RandomUint64())
	assert.NoError(t, tx.Sign(privKey))
	assert.NotNil(t, tx.Signature)
	assert.NoError(t, tx.Verify())

	return tx
}

func MarshallTestTransaction(t *testing.T, tx *Transaction) []byte {
	t.Helper()

	e, err := codec.EncodeProto(tx)
	assert.NoError(t, err)

	return e
}

func UnMarshallTestTransaction(t *testing.T, b []byte) *Transaction {
	t.Helper()

	d := new(Transaction)
	assert.NoError(t, codec.DecodeProto(b, d))

	return d
}

func GenerateRandomTestHeader(t *testing.T) *Header {
	t.Helper()

	return &Header{
		Version:       Version,
		MerkleRoot:    util.RandomHash(),
		PrevBlockHash: util.RandomHash(),
		Timestamp:     time.Now().UnixNano(),
		Height:        util.RandomUint64(),
	}
}

func GenerateRandomTestHeaderWithBody(t *testing.T, b *Body) *Header {
	t.Helper()

	m, err := b.CalculateMerkleRoot()
	assert.NoError(t, err)

	return &Header{
		Version:       Version,
		MerkleRoot:    m,
		PrevBlockHash: util.RandomHash(),
		Timestamp:     time.Now().UnixNano(),
		Height:        util.RandomUint64(),
	}
}

func GenerateRandomTestHeaderWithBodyAndHeight(t *testing.T, b *Body, h uint64) *Header {
	t.Helper()

	m, err := b.CalculateMerkleRoot()
	assert.NoError(t, err)

	return &Header{
		Version:       Version,
		MerkleRoot:    m,
		PrevBlockHash: util.RandomHash(),
		Timestamp:     time.Now().UnixNano(),
		Height:        h,
	}
}

func MarshallTestHeader(t *testing.T, h *Header) []byte {
	t.Helper()

	e, err := codec.EncodeProto(h)
	assert.NoError(t, err)

	return e
}

func UnmarshallTestHeader(t *testing.T, b []byte) *Header {
	t.Helper()

	d := new(Header)
	assert.NoError(t, codec.DecodeProto(b, d))

	return d
}

func GenerateRandomTestBody(t *testing.T, txCount int) *Body {
	t.Helper()

	txx := make([]*Transaction, txCount)
	for i := 0; i < txCount; i++ {
		txx[i] = GenerateRandomTestTransaction(t)
	}

	return NewBody(txx)
}

func MarshallTestBody(t *testing.T, body *Body) []byte {
	t.Helper()
	b, err := codec.EncodeProto(body)
	assert.NoError(t, err)
	return b
}

func UnmarshallTestBody(t *testing.T, b []byte) *Body {
	t.Helper()
	body := new(Body)
	assert.NoError(t, codec.DecodeProto(b, body))
	return body
}

func GenerateRandomBlock(t *testing.T, txCount int) *Block {
	t.Helper()

	body := GenerateRandomTestBody(t, txCount)
	header := GenerateRandomTestHeaderWithBody(t, body)

	block, err := NewBlock(header, body)
	assert.NoError(t, err)

	privKey, _ := crypto.GenerateTestKeyPair(t)

	assert.NoError(t, block.Sign(privKey))
	assert.NotNil(t, block.Signature)
	assert.NoError(t, block.Verify())

	return block
}

func GenerateRandomBlockWithHeight(t *testing.T, txCount int, h uint64) *Block {
	t.Helper()

	body := GenerateRandomTestBody(t, txCount)
	header := GenerateRandomTestHeaderWithBodyAndHeight(t, body, h)

	block, err := NewBlock(header, body)
	assert.NoError(t, err)

	privKey, _ := crypto.GenerateTestKeyPair(t)

	assert.NoError(t, block.Sign(privKey))
	assert.NotNil(t, block.Signature)
	assert.NoError(t, block.Verify())

	return block
}

func MarshallTestBlock(t *testing.T, block *Block) []byte {
	t.Helper()

	b, err := codec.EncodeProto(block)
	assert.NoError(t, err)

	return b
}

func UnmarshallTestBlock(t *testing.T, b []byte) *Block {
	t.Helper()

	block := new(Block)
	assert.NoError(t, codec.DecodeProto(b, block))

	return block
}
