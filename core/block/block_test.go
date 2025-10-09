package block

import (
	"github.com/andantan/p2p-pbft-modular-blockchain-network/codec"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBlock_SignAndVerify(t *testing.T) {
	block := GenerateRandomTestBlock(t, 1<<6)

	// Tempering proposer
	otherPrivKey, _ := crypto.GeneratePrivateKey()
	block.Proposer = otherPrivKey.PublicKey()
	assert.Error(t, block.Verify())
}

func TestBlock_EncodeDecode(t *testing.T) {
	block := GenerateRandomTestBlock(t, 1<<10)

	// Marshalling
	encodedBytes := MarshallTestBlock(t, block)
	// UnMarshalling
	decodedBlock := UnmarshallTestBlock(t, encodedBytes)

	hashOrig, _ := block.Hash()
	hashDecode, _ := decodedBlock.Hash()
	assert.True(t, hashOrig.Equal(hashDecode))
}

func TestNewBlockFromPrevHeader(t *testing.T) {
	prevHeader := GenerateRandomTestHeader(t)

	body := GenerateRandomTestBody(t, 1<<4)

	block, err := NewBlockFromPrevHeader(prevHeader, body)
	assert.NoError(t, err)

	assert.Equal(t, prevHeader.Height+1, block.Header.Height)

	prevBlockHash, _ := prevHeader.Hash()
	assert.True(t, prevBlockHash.Equal(block.Header.PrevBlockHash))

	merkleRoot, _ := body.Hash()
	assert.True(t, merkleRoot.Equal(block.Header.MerkleRoot))

	assert.Equal(t, body.GetWeight(), block.Header.Weight)

	blockHash, _ := block.Hash()
	assert.False(t, blockHash.IsZero())
}

func TestBlock_Seal(t *testing.T) {
	validatorCount := 4
	keys := make([]*crypto.PrivateKey, validatorCount)
	addrs := make([]types.Address, validatorCount)
	for i := 0; i < validatorCount; i++ {
		keys[i], _ = crypto.GeneratePrivateKey()
		addrs[i] = keys[i].PublicKey().Address()
	}
	quorum := (2 * validatorCount / 3) + 1

	t.Run("success and already sealed", func(t *testing.T) {
		b := GenerateRandomTestBlock(t, 1<<4)
		bh, err := b.Hash()
		assert.NoError(t, err)

		votes := make([]*CommitVote, quorum)
		for i := 0; i < quorum; i++ {
			votes[i] = GenerateRandomTestCommitVoteWithKey(t, bh.Bytes(), 0, b.Header.Height, keys[i])
		}

		assert.NoError(t, b.Seal(votes, addrs))
		assert.NotNil(t, b.Tail)
		assert.Equal(t, quorum, len(b.Tail.CommitVotes))

		err = b.Seal(votes, addrs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already set")
	})

	t.Run("error with not enough votes", func(t *testing.T) {
		b := GenerateRandomTestBlock(t, 1<<4)
		bh, err := b.Hash()
		assert.NoError(t, err)

		votes := make([]*CommitVote, quorum)
		for i := 0; i < quorum; i++ {
			votes[i] = GenerateRandomTestCommitVoteWithKey(t, bh.Bytes(), 0, b.Header.Height, keys[i])
		}

		invalidVotes := votes[:quorum-1]
		err = b.Seal(invalidVotes, addrs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not enough commit votes")
	})

	t.Run("error with non-validator", func(t *testing.T) {
		b := GenerateRandomTestBlock(t, 1<<4)
		bh, err := b.Hash()
		assert.NoError(t, err)

		votes := make([]*CommitVote, quorum)
		for i := 0; i < quorum; i++ {
			votes[i] = GenerateRandomTestCommitVoteWithKey(t, bh.Bytes(), 0, b.Header.Height, keys[i])
		}

		votes[0] = GenerateRandomTestCommitVote(t, bh.Bytes(), 0, b.Header.Height)
		err = b.Seal(votes, addrs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "non-validator")
	})
}

func TestBlock_EncodeDecode_WithTail(t *testing.T) {
	privKey, _ := crypto.GeneratePrivateKey()
	b := GenerateRandomTestBlock(t, 1<<4)
	bh, err := b.Hash()
	assert.NoError(t, err)
	assert.NoError(t, b.Sign(privKey))

	vote := GenerateRandomTestCommitVoteWithKey(t, bh.Bytes(), 0, b.Header.Height, privKey)
	validatorAddr := privKey.PublicKey().Address()
	assert.NoError(t, b.Seal([]*CommitVote{vote}, []types.Address{validatorAddr}))
	assert.NotNil(t, b.Tail)

	encodedBytes, err := codec.EncodeProto(b)
	assert.NoError(t, err)

	decodedBlock := new(Block)
	assert.NoError(t, codec.DecodeProto(encodedBytes, decodedBlock))

	hashOrig, _ := b.Hash()
	hashDecode, _ := decodedBlock.Hash()
	assert.True(t, hashOrig.Equal(hashDecode))

	assert.NotNil(t, decodedBlock.Tail)
	assert.Equal(t, 1, len(decodedBlock.Tail.CommitVotes))
	assert.True(t, b.Tail.CommitVotes[0].PublicKey.Equal(decodedBlock.Tail.CommitVotes[0].PublicKey))
}
