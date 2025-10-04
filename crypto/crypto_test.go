package crypto

import (
	"github.com/andantan/p2p-pbft-modular-blockchain-network/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGeneratePrivateKey(t *testing.T) {
	privKey, err := GeneratePrivateKey()
	assert.NoError(t, err)
	assert.NotNil(t, privKey.key)
}

func TestPublicKeyToAddress(t *testing.T) {
	privKey, err := GeneratePrivateKey()
	assert.NoError(t, err)

	pubKey := privKey.PublicKey()
	address := pubKey.Address()

	assert.Equal(t, types.AddressLength, len(address.Bytes()))
	assert.False(t, address.IsZero())
}

func TestSignAndVerify(t *testing.T) {
	data := []byte("hello world")

	privKey, err := GeneratePrivateKey()
	assert.NoError(t, err)

	pubKey := privKey.PublicKey()
	sig, err := privKey.Sign(data)
	assert.NoError(t, err)
	assert.True(t, sig.Verify(pubKey, data))

	assert.False(t, sig.Verify(pubKey, []byte("different data")))

	otherPrivKey, _ := GeneratePrivateKey()
	otherPubKey := otherPrivKey.PublicKey()
	assert.False(t, sig.Verify(otherPubKey, data))
}

func TestSignatureSerialization(t *testing.T) {
	data := []byte("hello world")

	privKey, err := GeneratePrivateKey()
	assert.NoError(t, err)
	
	sig, err := privKey.Sign(data)
	assert.NoError(t, err)

	sigBytes := sig.Bytes()
	recoveredSig, err := SignatureFromBytes(sigBytes)
	assert.NoError(t, err)
	assert.True(t, sig.R.Cmp(recoveredSig.R) == 0)
	assert.True(t, sig.S.Cmp(recoveredSig.S) == 0)

	sigString := sig.String()
	recoveredSigFromString, err := SignatureFromHexString(sigString)
	assert.NoError(t, err)
	assert.True(t, sig.R.Cmp(recoveredSigFromString.R) == 0)
	assert.True(t, sig.S.Cmp(recoveredSigFromString.S) == 0)
}
