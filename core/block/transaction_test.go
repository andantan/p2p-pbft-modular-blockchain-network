package block

import (
	"github.com/andantan/p2p-pbft-modular-blockchain-network/codec"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTransaction_SignAndVerify(t *testing.T) {
	privKey, _ := crypto.GeneratePrivateKey()
	tx := NewTransaction(util.RandomBytes(100), 0)

	assert.NoError(t, tx.Sign(privKey))
	assert.NotNil(t, tx.Signature)

	assert.NoError(t, tx.Verify())

	otherPrivKey, _ := crypto.GeneratePrivateKey()
	tx.From = otherPrivKey.PublicKey()
	assert.Error(t, tx.Verify())
}

func TestTransaction_EncodeDecode(t *testing.T) {
	tx := NewTransaction(util.RandomBytes(200), 1)
	privKey, _ := crypto.GeneratePrivateKey()
	assert.NoError(t, tx.Sign(privKey))

	// Marshalling
	b, err := codec.EncodeProto(tx)
	assert.NoError(t, err)

	// Unmarshalling
	txDecode := new(Transaction)
	assert.NoError(t, codec.DecodeProto(b, txDecode))

	hashOrig, _ := tx.Hash()
	hashDecode, _ := txDecode.Hash()
	assert.True(t, hashOrig.Eq(hashDecode))
	assert.Equal(t, tx.Data, txDecode.Data)
}

func TestTransaction_DataHash(t *testing.T) {
	tx := NewTransaction(util.RandomBytes(50), 0)

	hash := tx.dataHash()
	assert.False(t, hash.IsZero())
}

func TestTransaction_Hash(t *testing.T) {
	tx := NewTransaction(util.RandomBytes(50), 0)
	privKey, _ := crypto.GeneratePrivateKey()
	assert.NoError(t, tx.Sign(privKey))

	hash1, err := tx.Hash()
	assert.NoError(t, err)
	assert.False(t, hash1.IsZero())

	hash2, err := tx.Hash()
	assert.NoError(t, err)

	assert.True(t, hash1.Eq(hash2))
}
