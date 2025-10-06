package block

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBody_CalculateMerkleRoot(t *testing.T) {
	bodyWithOneTx := GenerateRandomTestBody(t, 1)
	root1, err := bodyWithOneTx.CalculateMerkleRoot()
	assert.NoError(t, err)
	tx1Hash, _ := bodyWithOneTx.Transactions[0].Hash()
	assert.True(t, root1.Eq(tx1Hash))

	bodyWithTwoTxs := GenerateRandomTestBody(t, 2)
	bodyWithReversedTxs := &Body{
		Transactions: []*Transaction{
			bodyWithTwoTxs.Transactions[1],
			bodyWithTwoTxs.Transactions[0],
		},
	}

	root2, err := bodyWithTwoTxs.CalculateMerkleRoot()
	assert.NoError(t, err)
	root3, err := bodyWithReversedTxs.CalculateMerkleRoot()
	assert.NoError(t, err)

	assert.False(t, root2.IsZero())
	assert.True(t, root2.Eq(root3))

	bodyEmpty := &Body{Transactions: []*Transaction{}}
	_, err = bodyEmpty.CalculateMerkleRoot()
	assert.Error(t, err)
}

func TestBody_EncodeDecode(t *testing.T) {
	originalBody := GenerateRandomTestBody(t, 5)

	// Marshalling
	encodedBytes := MarshallTestBody(t, originalBody)
	// UnMarshalling
	decodedBody := UnmarshallTestBody(t, encodedBytes)

	rootOrig, _ := originalBody.CalculateMerkleRoot()
	rootDecode, _ := decodedBody.CalculateMerkleRoot()
	assert.True(t, rootOrig.Eq(rootDecode))
	assert.Equal(t, len(originalBody.Transactions), len(decodedBody.Transactions))
}
