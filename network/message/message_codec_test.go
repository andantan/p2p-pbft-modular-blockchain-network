package message

import (
	"github.com/andantan/modular-blockchain/core/block"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaultGossipMessageCodec_EncodeDecode(t *testing.T) {
	c := NewDefaultGossipMessageCodec()

	t.Run("transaction_encode_decode", func(t *testing.T) {
		tx := block.GenerateRandomTestTransaction(t)
		encoded, err := c.Encode(tx)
		assert.NoError(t, err)
		assert.Equal(t, encoded[0], byte(MessageGossipType))
		assert.Equal(t, encoded[1], byte(MessageGossipSubTypeTransaction))

		decoded, err := c.Decode(encoded)
		assert.NoError(t, err)
		decodedTx, ok := decoded.(*block.Transaction)
		assert.True(t, ok)

		txh, err := tx.Hash()
		assert.NoError(t, err)
		decodedTxh, err := decodedTx.Hash()
		assert.NoError(t, err)
		assert.True(t, txh.Equal(decodedTxh))
	})

	t.Run("block_encode_decode", func(t *testing.T) {
		b := block.GenerateRandomTestBlock(t, 1<<3)
		encoded, err := c.Encode(b)
		assert.NoError(t, err)
		assert.Equal(t, encoded[0], byte(MessageGossipType))
		assert.Equal(t, encoded[1], byte(MessageGossipSubTypeBlock))

		decoded, err := c.Decode(encoded)
		assert.NoError(t, err)
		decodedB, ok := decoded.(*block.Block)
		assert.True(t, ok)

		bh, err := b.Hash()
		assert.NoError(t, err)
		decodedBh, err := decodedB.Hash()
		assert.NoError(t, err)
		assert.True(t, bh.Equal(decodedBh))
	})
}
