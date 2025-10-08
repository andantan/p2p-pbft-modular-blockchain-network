package tcp

import (
	"github.com/andantan/p2p-pbft-modular-blockchain-network/codec"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHandshakeMessage_SignVerify(t *testing.T) {
	msg, _ := GenerateTestTCPHandshakeMessage(t, "127.0.0.1:4000")

	otherPrivKey, _ := crypto.GeneratePrivateKey()
	msg.PublicKey = otherPrivKey.PublicKey()
	assert.Error(t, msg.Verify())
}

func TestHandshakeMessage_EncodeDecode(t *testing.T) {
	originMsg, _ := GenerateTestTCPHandshakeMessage(t, "127.0.0.1:4000")

	// Marshalling
	encodedBytes, err := codec.EncodeProto(originMsg)
	assert.NoError(t, err)

	// UnMarshalling
	decodedMsg := new(TCPHandshakeMessage)
	assert.NoError(t, codec.DecodeProto(encodedBytes, decodedMsg))

	assert.NoError(t, decodedMsg.Verify())

	originHash, err := originMsg.Hash()
	assert.NoError(t, err)
	decodedHash, err := decodedMsg.Hash()
	assert.NoError(t, err)

	assert.Equal(t, originHash, decodedHash)
}
