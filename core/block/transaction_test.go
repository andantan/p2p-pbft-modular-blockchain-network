package block

import (
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTransaction_SignAndVerify(t *testing.T) {
	tx := GenerateRandomTestTransaction(t)

	otherPrivKey, _ := crypto.GeneratePrivateKey()
	tx.From = otherPrivKey.PublicKey()
	assert.Error(t, tx.Verify())
}

func TestTransaction_EncodeDecode(t *testing.T) {
	tx := GenerateRandomTestTransaction(t)

	// Marshalling
	et := MarshallTestTransaction(t, tx)
	// Unmarshalling
	dt := UnMarshallTestTransaction(t, et)

	hashOrig, _ := tx.Hash()
	hashDecode, _ := dt.Hash()
	assert.True(t, hashOrig.Equal(hashDecode))
	assert.Equal(t, tx.Data, dt.Data)
}

func TestTransaction_DataHash(t *testing.T) {
	tx := GenerateRandomTestTransaction(t)

	hash := tx.dataHash()
	assert.False(t, hash.IsZero())
}

func TestTransaction_Hash(t *testing.T) {
	tx := GenerateRandomTestTransaction(t)

	hash1, err := tx.Hash()
	assert.NoError(t, err)
	assert.False(t, hash1.IsZero())

	hash2, err := tx.Hash()
	assert.NoError(t, err)

	assert.True(t, hash1.Equal(hash2))
}
