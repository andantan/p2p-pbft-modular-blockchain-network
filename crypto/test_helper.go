package crypto

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func GenerateTestKeyPair(t *testing.T) (*PrivateKey, *PublicKey) {
	t.Helper()

	privKey, err := GeneratePrivateKey()
	assert.NoError(t, err)
	assert.NotNil(t, privKey)
	pubKey := privKey.PublicKey()
	assert.NotNil(t, pubKey)

	return privKey, pubKey
}
