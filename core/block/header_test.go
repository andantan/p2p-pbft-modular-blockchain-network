package block

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHeader_EncodeDecode(t *testing.T) {
	h := GenerateRandomTestHeader(t)

	// Marshalling
	eh := MarshallTestHeader(t, h)
	// Unmarshalling
	dh := UnmarshallTestHeader(t, eh)

	assert.Equal(t, *h, *dh)
}

func TestHeader_Hash(t *testing.T) {
	h1 := GenerateRandomTestHeader(t)
	h2 := *h1

	hash1, err := h1.Hash()
	assert.NoError(t, err)
	hash2, err := (&h2).Hash()
	assert.NoError(t, err)

	assert.True(t, hash1.Equal(hash2))

	h2.Version = 2
	hash3, err := (&h2).Hash()
	assert.NoError(t, err)
	assert.False(t, hash1.Equal(hash3))
}
