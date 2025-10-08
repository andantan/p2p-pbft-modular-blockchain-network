package pbft

import (
	"github.com/andantan/p2p-pbft-modular-blockchain-network/codec"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPbftPrePrepareMessage_SignVerify(t *testing.T) {
	msg, _ := GenerateTestPbftPrePrepareMessage(t, 0)

	otherPrivKey, _ := crypto.GeneratePrivateKey()
	msg.PublicKey = otherPrivKey.PublicKey()
	assert.Error(t, msg.Verify())
}

func TestPbftPrePrepareMessage_EncodeDecode(t *testing.T) {
	msg, _ := GenerateTestPbftPrePrepareMessage(t, 0)

	// Marshalling
	encoded, err := codec.EncodeProto(msg)
	assert.NoError(t, err)

	// UnMarshalling
	decodedMsg := new(PbftPrePrepareMessage)
	assert.NoError(t, codec.DecodeProto(encoded, decodedMsg))

	originHash, err := msg.Hash()
	assert.NoError(t, err)
	decodedHash, err := decodedMsg.Hash()
	assert.NoError(t, err)
	assert.Equal(t, originHash, decodedHash)
}

func TestPbftPrepareMessage_SignVerify(t *testing.T) {
	msg, _ := GenerateTestPbftPrepareMessage(t, 1)

	otherPrivKey, _ := crypto.GeneratePrivateKey()
	msg.PublicKey = otherPrivKey.PublicKey()
	assert.Error(t, msg.Verify())
}

func TestPbftPrepareMessage_EncodeDecode(t *testing.T) {
	msg, _ := GenerateTestPbftPrepareMessage(t, 1)

	encoded, err := codec.EncodeProto(msg)
	assert.NoError(t, err)

	decodedMsg := new(PbftPrepareMessage)
	assert.NoError(t, codec.DecodeProto(encoded, decodedMsg))

	originHash, err := msg.Hash()
	assert.NoError(t, err)
	decodedHash, err := decodedMsg.Hash()
	assert.NoError(t, err)
	assert.Equal(t, originHash, decodedHash)
}

func TestPbftCommitMessage_SignVerify(t *testing.T) {
	msg, _ := GenerateTestPbftCommitMessage(t, 0)

	otherPrivKey, _ := crypto.GeneratePrivateKey()
	msg.PublicKey = otherPrivKey.PublicKey()
	assert.Error(t, msg.Verify())
}

func TestPbftCommitMessage_EncodeDecode(t *testing.T) {
	msg, _ := GenerateTestPbftCommitMessage(t, 1)

	encoded, err := codec.EncodeProto(msg)
	assert.NoError(t, err)

	decodedMsg := new(PbftCommitMessage)
	assert.NoError(t, codec.DecodeProto(encoded, decodedMsg))

	originHash, err := msg.Hash()
	assert.NoError(t, err)
	decodedHash, err := decodedMsg.Hash()
	assert.NoError(t, err)
	assert.Equal(t, originHash, decodedHash)
}

func TestPbftViewChangeMessage_SignVerify(t *testing.T) {
	msg, _ := GenerateTestPbftViewChangeMessage(t, 1, 10)

	otherKey, _ := crypto.GeneratePrivateKey()
	msg.PublicKey = otherKey.PublicKey()
	assert.Error(t, msg.Verify())
}

func TestPbftViewChangeMessage_EncodeDecode(t *testing.T) {
	msg, _ := GenerateTestPbftViewChangeMessage(t, 1, 10)

	encoded, err := codec.EncodeProto(msg)
	assert.NoError(t, err)

	decodedMsg := new(PbftViewChangeMessage)
	assert.NoError(t, codec.DecodeProto(encoded, decodedMsg))

	originHash, err := msg.Hash()
	assert.NoError(t, err)
	decodedHash, err := decodedMsg.Hash()
	assert.NoError(t, err)
	assert.Equal(t, originHash, decodedHash)
}

func TestPbftNewViewMessage_SignVerify(t *testing.T) {
	msg, _ := GenerateTestPbftNewViewMessage(t, 1, 10, 10)

	otherKey, _ := crypto.GeneratePrivateKey()
	msg.PublicKey = otherKey.PublicKey()
	assert.Error(t, msg.Verify())
}

func TestPbftNewViewMessage_EncodeDecode(t *testing.T) {
	msg, _ := GenerateTestPbftNewViewMessage(t, 1, 10, 8)

	encoded, err := codec.EncodeProto(msg)
	assert.NoError(t, err)

	decodedMsg := new(PbftNewViewMessage)
	assert.NoError(t, codec.DecodeProto(encoded, decodedMsg))

	assert.Equal(t, msg.NewView, decodedMsg.NewView)
	assert.Equal(t, len(msg.ViewChangeMessages), len(decodedMsg.ViewChangeMessages))
	assert.NotNil(t, decodedMsg.PrePrepareMessage)
}
