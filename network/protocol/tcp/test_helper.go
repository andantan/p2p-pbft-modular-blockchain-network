package tcp

import (
	"github.com/andantan/modular-blockchain/crypto"
	"github.com/stretchr/testify/assert"
	"testing"
)

func GenerateTestTCPHandshakeMessage(t *testing.T, addr string) (*TcpHandshakeMessage, *crypto.PrivateKey) {
	t.Helper()

	privKey, _ := crypto.GeneratePrivateKey()
	msg := NewTcpHandshakeMessage(privKey.PublicKey(), addr)
	assert.NoError(t, msg.Sign(privKey))
	assert.NotNil(t, msg.Signature)
	assert.NoError(t, msg.Verify())

	return msg, privKey
}

func GenerateTestTcpNode(t *testing.T, listenAddr string) *TcpNode {
	t.Helper()

	privKeyA, err := crypto.GeneratePrivateKey()
	assert.NoError(t, err)

	return NewTcpNode(privKeyA, listenAddr)
}
