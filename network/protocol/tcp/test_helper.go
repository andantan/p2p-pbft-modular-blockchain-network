package tcp

import (
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
	"github.com/stretchr/testify/assert"
	"testing"
)

func GenerateTestTCPHandshakeMessage(t *testing.T, addr string) (*TCPHandshakeMessage, *crypto.PrivateKey) {
	t.Helper()

	privKey, _ := crypto.GeneratePrivateKey()
	msg := NewTCPHandshakeMessage(privKey.PublicKey(), addr)
	assert.NoError(t, msg.Sign(privKey))
	assert.NotNil(t, msg.Signature)
	assert.NoError(t, msg.Verify())

	return msg, privKey
}
