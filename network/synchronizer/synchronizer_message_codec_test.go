package synchronizer

import (
	"github.com/andantan/modular-blockchain/core/block"
	"github.com/andantan/modular-blockchain/network/message"
	"github.com/andantan/modular-blockchain/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaultSyncMessageCodec_EncodeDecode(t *testing.T) {
	c := NewDefaultSyncMessageCodec()

	t.Run("request_status_message_encode_decode", func(t *testing.T) {
		req := &RequestStatusMessage{}
		encoded, err := c.Encode(req)
		assert.NoError(t, err)
		assert.Equal(t, encoded[0], byte(message.MessageSyncType))
		assert.Equal(t, encoded[1], byte(MessageSyncSubTypeRequestStatus))

		decoded, err := c.Decode(encoded)
		assert.NoError(t, err)

		_, ok := decoded.(*RequestStatusMessage)
		assert.True(t, ok)
	})

	t.Run("response_status_message_encode_decode", func(t *testing.T) {
		addr := util.RandomAddress()
		res := &ResponseStatusMessage{
			Address:          addr,
			NetAddr:          "0.0.0.0:8080",
			Version:          uint32(2),
			Height:           uint64(55),
			State:            byte(1),
			GenesisBlockHash: util.RandomHash(),
			CurrentBlockHash: util.RandomHash(),
		}

		encoded, err := c.Encode(res)
		assert.NoError(t, err)
		assert.Equal(t, encoded[0], byte(message.MessageSyncType))
		assert.Equal(t, encoded[1], byte(MessageSyncSubTypeResponseStatus))

		decoded, err := c.Decode(encoded)
		assert.NoError(t, err)

		decodedRes, ok := decoded.(*ResponseStatusMessage)
		assert.True(t, ok)

		assert.True(t, res.Address.Equal(decodedRes.Address))
		assert.Equal(t, res.NetAddr, decodedRes.NetAddr)
		assert.Equal(t, res.Version, decodedRes.Version)
		assert.Equal(t, res.Height, decodedRes.Height)
		assert.Equal(t, res.State, decodedRes.State)
		assert.True(t, res.GenesisBlockHash.Equal(decodedRes.GenesisBlockHash))
		assert.True(t, res.CurrentBlockHash.Equal(decodedRes.CurrentBlockHash))
	})

	t.Run("request_headers_message_encode_decode", func(t *testing.T) {
		req := &RequestHeadersMessage{
			From:  uint64(55),
			Count: uint64(11),
		}

		encoded, err := c.Encode(req)
		assert.NoError(t, err)
		assert.Equal(t, encoded[0], byte(message.MessageSyncType))
		assert.Equal(t, encoded[1], byte(MessageSyncSubTypeRequestHeaders))

		decoded, err := c.Decode(encoded)
		assert.NoError(t, err)

		decodedReq, ok := decoded.(*RequestHeadersMessage)
		assert.True(t, ok)

		assert.Equal(t, req.From, decodedReq.From)
		assert.Equal(t, req.Count, decodedReq.Count)
	})

	t.Run("response_headers_message_encode_decode", func(t *testing.T) {
		headers := []*block.Header{
			block.GenerateRandomTestHeader(t),
			block.GenerateRandomTestHeader(t),
			block.GenerateRandomTestHeader(t),
			block.GenerateRandomTestHeader(t),
			block.GenerateRandomTestHeader(t),
		}
		res := &ResponseHeadersMessage{Headers: headers}

		encoded, err := c.Encode(res)
		assert.NoError(t, err)
		assert.Equal(t, encoded[0], byte(message.MessageSyncType))
		assert.Equal(t, encoded[1], byte(MessageSyncSubTypeResponseHeaders))

		decoded, err := c.Decode(encoded)
		assert.NoError(t, err)

		decodedRes, ok := decoded.(*ResponseHeadersMessage)
		assert.True(t, ok)

		for i, h := range headers {
			hash, err := h.Hash()
			assert.NoError(t, err)
			decodedHash, err := decodedRes.Headers[i].Hash()
			assert.NoError(t, err)
			assert.True(t, decodedHash.Equal(hash))
		}
	})

	t.Run("request_block_message_encode_decode", func(t *testing.T) {
		req := &RequestBlocksMessage{
			From:  uint64(55),
			Count: uint64(11),
		}

		encoded, err := c.Encode(req)
		assert.NoError(t, err)
		assert.Equal(t, encoded[0], byte(message.MessageSyncType))
		assert.Equal(t, encoded[1], byte(MessageSyncSubTypeRequestBlocks))

		decoded, err := c.Decode(encoded)
		assert.NoError(t, err)
		decodedReq, ok := decoded.(*RequestBlocksMessage)
		assert.True(t, ok)

		assert.Equal(t, req.From, decodedReq.From)
		assert.Equal(t, req.Count, decodedReq.Count)
	})

	t.Run("response_block_message_encode_decode", func(t *testing.T) {
		blocks := []*block.Block{
			block.GenerateRandomTestBlock(t, 1<<3),
			block.GenerateRandomTestBlock(t, 1<<6),
			block.GenerateRandomTestBlock(t, 1<<11),
			block.GenerateRandomTestBlock(t, 1<<2),
			block.GenerateRandomTestBlock(t, 1<<7),
		}
		res := &ResponseBlocksMessage{Blocks: blocks}

		encoded, err := c.Encode(res)
		assert.NoError(t, err)
		assert.Equal(t, encoded[0], byte(message.MessageSyncType))
		assert.Equal(t, encoded[1], byte(MessageSyncSubTypeResponseBlocks))

		decoded, err := c.Decode(encoded)
		assert.NoError(t, err)

		decodedRes, ok := decoded.(*ResponseBlocksMessage)
		assert.True(t, ok)

		for i, b := range blocks {
			hash, err := b.Hash()
			assert.NoError(t, err)
			decodedHash, err := decodedRes.Blocks[i].Hash()
			assert.NoError(t, err)
			assert.True(t, decodedHash.Equal(hash))
		}
	})
}
