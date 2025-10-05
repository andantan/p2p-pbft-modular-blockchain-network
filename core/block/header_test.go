package block

import (
	"github.com/andantan/p2p-pbft-modular-blockchain-network/codec"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/util"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestHeader_EncodeDecode(t *testing.T) {
	h := &Header{
		Version:       1,
		MerkleRoot:    util.RandomHash(),
		PrevBlockHash: util.RandomHash(),
		Timestamp:     time.Now().UnixNano(),
		Height:        10,
	}

	// Marshalling
	b, err := codec.EncodeProto(h)
	assert.NoError(t, err)

	// Unmarshalling
	hDecode := new(Header)
	err = codec.DecodeProto(b, hDecode)
	assert.NoError(t, err)

	assert.Equal(t, h, hDecode)
}

func TestHeader_Hash(t *testing.T) {
	h1 := &Header{
		Version:       1,
		PrevBlockHash: util.RandomHash(),
		Height:        1,
	}
	h2 := *h1

	hash1, err := h1.Hash()
	assert.NoError(t, err)
	hash2, err := (&h2).Hash()
	assert.NoError(t, err)

	assert.True(t, hash1.Eq(hash2))

	h2.Version = 2
	hash3, err := (&h2).Hash()
	assert.NoError(t, err)
	assert.False(t, hash1.Eq(hash3))
}
