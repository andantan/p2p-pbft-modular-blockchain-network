package types

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAddress_BytesAndFromBytes(t *testing.T) {
	originalBytes := make([]byte, AddressLength)
	originalBytes[0] = 0xaa
	originalBytes[19] = 0xbb

	// FromBytes
	addr, err := AddressFromBytes(originalBytes)
	assert.NoError(t, err)

	// Bytes
	resultBytes := addr.Bytes()
	assert.Equal(t, originalBytes, resultBytes)

	_, err = AddressFromBytes([]byte{1, 2, 3})
	assert.Error(t, err)
}

func TestAddress_IsZero(t *testing.T) {
	zeroAddr := Address{}
	assert.True(t, zeroAddr.IsZero())

	nonZeroBytes := make([]byte, AddressLength)
	nonZeroBytes[10] = 0x01
	nonZeroAddr, _ := AddressFromBytes(nonZeroBytes)
	assert.False(t, nonZeroAddr.IsZero())
}
