package message

import (
	"github.com/andantan/modular-blockchain/codec"
	"github.com/andantan/modular-blockchain/core/block"
	"github.com/andantan/modular-blockchain/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSyncMessages_EncodeDecode(t *testing.T) {
	t.Run("StatusMessages", func(t *testing.T) {
		// Request
		req := &RequestStatusMessage{}
		encodedReq, err := codec.EncodeProto(req)
		assert.NoError(t, err)
		decodedReq := new(RequestStatusMessage)
		assert.NoError(t, codec.DecodeProto(encodedReq, decodedReq))

		// Response
		resp := &ResponseStatusMessage{
			Address:          util.RandomAddress(),
			Version:          1,
			Height:           100,
			GenesisBlockHash: util.RandomHash(),
		}
		encodedResp, err := codec.EncodeProto(resp)
		assert.NoError(t, err)
		decodedResp := new(ResponseStatusMessage)
		assert.NoError(t, codec.DecodeProto(encodedResp, decodedResp))
		assert.Equal(t, resp.Height, decodedResp.Height)
		assert.True(t, resp.Address.Equal(decodedResp.Address))
	})

	t.Run("HeadersMessages", func(t *testing.T) {
		// Request
		req := &RequestHeadersMessage{From: 10, Count: 5}
		encodedReq, err := codec.EncodeProto(req)
		assert.NoError(t, err)
		decodedReq := new(RequestHeadersMessage)
		assert.NoError(t, codec.DecodeProto(encodedReq, decodedReq))
		assert.Equal(t, req.From, decodedReq.From)
		assert.Equal(t, req.Count, decodedReq.Count)

		// Response
		resp := &ResponseHeadersMessage{
			Headers: []*block.Header{
				block.GenerateRandomTestHeader(t),
				block.GenerateRandomTestHeader(t),
			},
		}
		encodedResp, err := codec.EncodeProto(resp)
		assert.NoError(t, err)
		decodedResp := new(ResponseHeadersMessage)
		assert.NoError(t, codec.DecodeProto(encodedResp, decodedResp))
		assert.Equal(t, len(resp.Headers), len(decodedResp.Headers))
	})

	t.Run("BlocksMessages", func(t *testing.T) {
		// Request
		req := &RequestBlocksMessage{From: 20, Count: 2}
		encodedReq, err := codec.EncodeProto(req)
		assert.NoError(t, err)
		decodedReq := new(RequestBlocksMessage)
		assert.NoError(t, codec.DecodeProto(encodedReq, decodedReq))
		assert.Equal(t, req.Count, decodedReq.Count)

		// Response
		resp := &ResponseBlocksMessage{
			Blocks: []*block.Block{
				block.GenerateRandomTestBlock(t, 1<<4),
				block.GenerateRandomTestBlock(t, 1<<6),
			},
		}
		encodedResp, err := codec.EncodeProto(resp)
		assert.NoError(t, err)
		decodedResp := new(ResponseBlocksMessage)
		assert.NoError(t, codec.DecodeProto(encodedResp, decodedResp))
		assert.Equal(t, len(resp.Blocks), len(decodedResp.Blocks))
	})
}
