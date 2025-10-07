package block

import (
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
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
	assert.True(t, hashOrig.Eq(hashDecode))
}

func TestNewBlockFromPrevHeader(t *testing.T) {
	prevHeader := GenerateRandomTestHeader(t)

	body := GenerateRandomTestBody(t, 1<<4)

	block, err := NewBlockFromPrevHeader(prevHeader, body)
	assert.NoError(t, err)

	assert.Equal(t, prevHeader.Height+1, block.Header.Height)

	prevBlockHash, _ := prevHeader.Hash()
	assert.True(t, prevBlockHash.Eq(block.Header.PrevBlockHash))

	merkleRoot, _ := body.CalculateMerkleRoot()
	assert.True(t, merkleRoot.Eq(block.Header.MerkleRoot))

	assert.Equal(t, body.GetWeight(), block.Header.Weight)

	blockHash, _ := block.Hash()
	assert.False(t, blockHash.IsZero())
}
