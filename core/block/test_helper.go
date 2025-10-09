package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
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

	m, err := b.Hash()
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

	m, err := b.Hash()
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

func GenerateRandomTestCommitVote(t *testing.T, data []byte, view, sequence uint64) *CommitVote {
	t.Helper()

	buf := new(bytes.Buffer)
	assert.NoError(t, binary.Write(buf, binary.LittleEndian, view))
	assert.NoError(t, binary.Write(buf, binary.LittleEndian, sequence))
	buf.Write(data)

	digest := sha256.Sum256(buf.Bytes())

	privKey, pubKey := crypto.GenerateTestKeyPair(t)
	sig, err := privKey.Sign(digest[:])
	assert.NoError(t, err)
	assert.True(t, sig.Verify(pubKey, digest[:]))

	return NewCommitVote(digest, pubKey, sig)
}

func GenerateRandomTestCommitVoteWithKey(t *testing.T, data []byte, view, sequence uint64, k *crypto.PrivateKey) *CommitVote {
	t.Helper()

	buf := new(bytes.Buffer)
	assert.NoError(t, binary.Write(buf, binary.LittleEndian, view))
	assert.NoError(t, binary.Write(buf, binary.LittleEndian, sequence))
	buf.Write(data)

	digest := sha256.Sum256(buf.Bytes())

	sig, err := k.Sign(digest[:])
	assert.NoError(t, err)
	assert.True(t, sig.Verify(k.PublicKey(), digest[:]))

	return NewCommitVote(digest, k.PublicKey(), sig)
}

func MarshallTestCommitVote(t *testing.T, tail *CommitVote) []byte {
	t.Helper()

	b, err := codec.EncodeProto(tail)
	assert.NoError(t, err)

	return b
}

func UnMarshallTestCommitVote(t *testing.T, b []byte) *CommitVote {
	t.Helper()

	d := new(CommitVote)
	assert.NoError(t, codec.DecodeProto(b, d))
	return d
}

func GenerateRandomTestTail(t *testing.T, data []byte, n int, view, sequence uint64) *Tail {
	t.Helper()

	cvs := make([]*CommitVote, n)
	for i := 0; i < n; i++ {
		cvs[i] = GenerateRandomTestCommitVote(t, data, view, sequence)
	}

	return NewTail(cvs)
}

func MarshallTestTail(t *testing.T, tail *Tail) []byte {
	t.Helper()

	b, err := codec.EncodeProto(tail)
	assert.NoError(t, err)

	return b
}

func UnMarshallTestTail(t *testing.T, b []byte) *Tail {
	t.Helper()

	d := new(Tail)
	assert.NoError(t, codec.DecodeProto(b, d))
	return d
}

func GenerateRandomTestBlock(t *testing.T, txCount int) *Block {
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

func GenerateRandomTestBlockWithHeight(t *testing.T, txCount int, h uint64) *Block {
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

func GenerateRandomTestBlockWithPrevHeader(t *testing.T, h *Header, txCount int) *Block {
	t.Helper()

	bd := GenerateRandomTestBody(t, txCount)
	b, err := NewBlockFromPrevHeader(h, bd)
	assert.NoError(t, err)

	privKey, _ := crypto.GenerateTestKeyPair(t)
	assert.NoError(t, b.Sign(privKey))
	assert.NotNil(t, b.Signature)
	assert.NoError(t, b.Verify())

	return b
}

func GenerateRandomTestBlockWithPrevHeaderAndKey(t *testing.T, h *Header, txCount int, k *crypto.PrivateKey) *Block {
	t.Helper()

	bd := GenerateRandomTestBody(t, txCount)
	b, err := NewBlockFromPrevHeader(h, bd)
	assert.NoError(t, err)
	assert.NoError(t, b.Sign(k))
	assert.NotNil(t, b.Signature)
	assert.NoError(t, b.Verify())

	return b
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
