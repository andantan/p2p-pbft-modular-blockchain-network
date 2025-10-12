package pbft

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPbftConsensusMessageCodec_EncodeDecode(t *testing.T) {
	codec := NewPbftConsensusMessageCodec()

	t.Run("PrePrepareMessage", func(t *testing.T) {
		originalMsg, _ := GenerateTestPbftPrePrepareMessage(t, 0)
		encodedBytes, err := codec.Encode(originalMsg)
		assert.NoError(t, err)
		decodedMsg, err := codec.Decode(encodedBytes)
		assert.NoError(t, err)
		assert.IsType(t, &PbftPrePrepareMessage{}, decodedMsg)
		assert.Equal(t, originalMsg.Sequence, decodedMsg.(*PbftPrePrepareMessage).Sequence)
	})

	// --- Prepare ---
	t.Run("PrepareMessage", func(t *testing.T) {
		originalMsg, _ := GenerateTestPbftPrepareMessage(t, 0)
		encodedBytes, err := codec.Encode(originalMsg)
		assert.NoError(t, err)
		decodedMsg, err := codec.Decode(encodedBytes)
		assert.NoError(t, err)
		assert.IsType(t, &PbftPrepareMessage{}, decodedMsg)
		assert.True(t, originalMsg.BlockHash.Equal(decodedMsg.(*PbftPrepareMessage).BlockHash))
	})

	// --- Commit ---
	t.Run("CommitMessage", func(t *testing.T) {
		originalMsg, _ := GenerateTestPbftCommitMessage(t, 0)
		encodedBytes, err := codec.Encode(originalMsg)
		assert.NoError(t, err)
		decodedMsg, err := codec.Decode(encodedBytes)
		assert.NoError(t, err)
		assert.IsType(t, &PbftCommitMessage{}, decodedMsg)
		assert.True(t, originalMsg.BlockHash.Equal(decodedMsg.(*PbftCommitMessage).BlockHash))
	})

	// --- ViewChange ---
	t.Run("ViewChangeMessage", func(t *testing.T) {
		originalMsg, _ := GenerateTestPbftViewChangeMessage(t, 1, 10)
		encodedBytes, err := codec.Encode(originalMsg)
		assert.NoError(t, err)
		decodedMsg, err := codec.Decode(encodedBytes)
		assert.NoError(t, err)
		assert.IsType(t, &PbftViewChangeMessage{}, decodedMsg)
		assert.Equal(t, originalMsg.NewView, decodedMsg.(*PbftViewChangeMessage).NewView)
	})

	// --- NewView ---
	t.Run("NewViewMessage", func(t *testing.T) {
		originalMsg, _ := GenerateTestPbftNewViewMessage(t, 1, 10, 4)
		encodedBytes, err := codec.Encode(originalMsg)
		assert.NoError(t, err)
		decodedMsg, err := codec.Decode(encodedBytes)
		assert.NoError(t, err)
		assert.IsType(t, &PbftNewViewMessage{}, decodedMsg)
		assert.Equal(t, originalMsg.NewView, decodedMsg.(*PbftNewViewMessage).NewView)
	})
}
