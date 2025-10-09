package block

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCommitVote_EncodeDecode(t *testing.T) {
	data := []byte("hello")
	originalVote := GenerateRandomTestCommitVote(t, data)

	// Marshalling
	encodedBytes := MarshallTestCommitVote(t, originalVote)

	// Unmarshalling
	decodedVote := UnMarshallTestCommitVote(t, encodedBytes)

	assert.True(t, decodedVote.Signature.Verify(decodedVote.PublicKey, data))
	assert.True(t, originalVote.PublicKey.Equal(decodedVote.PublicKey))
	assert.True(t, originalVote.Signature.R.Cmp(decodedVote.Signature.R) == 0)
	assert.True(t, originalVote.Signature.S.Cmp(decodedVote.Signature.S) == 0)
}

func TestTail_EncodeDecode(t *testing.T) {
	data := []byte("hello")
	originalTail := GenerateRandomTestTail(t, data, 30)

	// Marshalling
	encodedBytes := MarshallTestTail(t, originalTail)

	// Unmarshalling
	decodedTail := UnMarshallTestTail(t, encodedBytes)

	assert.Equal(t, len(originalTail.CommitVotes), len(decodedTail.CommitVotes))

	for i := 0; i < len(originalTail.CommitVotes); i++ {
		originalVote := originalTail.CommitVotes[i]
		decodedVote := decodedTail.CommitVotes[i]

		assert.True(t, originalVote.PublicKey.Equal(decodedVote.PublicKey))
		assert.True(t, originalVote.Signature.R.Cmp(decodedVote.Signature.R) == 0)
		assert.True(t, originalVote.Signature.S.Cmp(decodedVote.Signature.S) == 0)
	}
}
