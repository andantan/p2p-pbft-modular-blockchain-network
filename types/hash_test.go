package types

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHash_BytesAndFromBytes(t *testing.T) {
	originalBytes := make([]byte, HashLength)
	originalBytes[0] = 0xde
	originalBytes[1] = 0xad
	originalBytes[31] = 0xef

	// FromBytes
	hash, err := HashFromBytes(originalBytes)
	assert.NoError(t, err)

	// Bytes
	resultBytes := hash.Bytes()
	assert.Equal(t, originalBytes, resultBytes)

	_, err = HashFromBytes([]byte{1, 2, 3})
	assert.Error(t, err)
}

func TestHash_StringAndFromHexString(t *testing.T) {
	originalStr := "deadbeef000000000000000000000000000000000000000000000000beefdead"

	// FromHexString
	hash, err := HashFromHexString(originalStr)
	assert.NoError(t, err)

	// String
	resultStr := hash.String()
	assert.Equal(t, "0x"+originalStr, resultStr)

	// FromHexString
	hashWithPrefix, err := HashFromHexString("0x" + originalStr)
	assert.NoError(t, err)
	assert.Equal(t, hash, hashWithPrefix)

	_, err = HashFromHexString("123456")
	assert.Error(t, err)

	_, err = HashFromHexString("gg")
	assert.Error(t, err)
}

func TestHash_IsZero(t *testing.T) {
	zeroHash := Hash{}
	assert.True(t, zeroHash.IsZero())

	nonZeroBytes := make([]byte, HashLength)
	nonZeroBytes[5] = 0xff
	nonZeroHash, _ := HashFromBytes(nonZeroBytes)
	assert.False(t, nonZeroHash.IsZero())
}
