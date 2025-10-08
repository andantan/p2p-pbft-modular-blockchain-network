package message

import (
	"github.com/andantan/p2p-pbft-modular-blockchain-network/core/block"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
	"github.com/stretchr/testify/assert"
	"testing"
)

func GenerateTestHandshakeMessage(t *testing.T, addr string) (*HandshakeMessage, *crypto.PrivateKey) {
	t.Helper()

	privKey, _ := crypto.GeneratePrivateKey()
	msg := NewHandshakeMessage(privKey.PublicKey(), addr)
	assert.NoError(t, msg.Sign(privKey))
	assert.NotNil(t, msg.Signature)
	assert.NoError(t, msg.Verify())

	return msg, privKey
}

func GenerateTestPbftPrePrepareMessage(t *testing.T, view uint64) (*PbftPrePrepareMessage, *crypto.PrivateKey) {
	t.Helper()

	privKey, _ := crypto.GeneratePrivateKey()
	b := block.GenerateRandomTestBlock(t, 1<<4)
	msg := NewPbftPrePrepareMessage(view, b.Header.Height, b, privKey.PublicKey())

	assert.NoError(t, msg.Sign(privKey))
	assert.NotNil(t, msg.Signature)
	assert.NoError(t, msg.Verify())

	return msg, privKey
}

func GenerateTestPbftPrepareMessage(t *testing.T, view uint64) (*PbftPrepareMessage, *crypto.PrivateKey) {
	t.Helper()

	privKey, _ := crypto.GeneratePrivateKey()
	b := block.GenerateRandomTestBlock(t, 1<<4)
	h, err := b.Hash()
	assert.NoError(t, err)
	msg := NewPbftPrepareMessage(view, b.Header.Height, h)

	assert.NoError(t, msg.Sign(privKey))
	assert.NotNil(t, msg.PublicKey)
	assert.NotNil(t, msg.Signature)
	assert.NoError(t, msg.Verify())

	return msg, privKey
}

func GenerateTestPbftCommitMessage(t *testing.T, view uint64) (*PbftCommitMessage, *crypto.PrivateKey) {
	t.Helper()

	privKey, _ := crypto.GeneratePrivateKey()
	b := block.GenerateRandomTestBlock(t, 1<<4)
	h, err := b.Hash()
	assert.NoError(t, err)
	msg := NewPbftCommitMessage(view, b.Header.Height, h)

	assert.NoError(t, msg.Sign(privKey))
	assert.NotNil(t, msg.PublicKey)
	assert.NotNil(t, msg.Signature)
	assert.NoError(t, msg.Verify())

	return msg, privKey
}
