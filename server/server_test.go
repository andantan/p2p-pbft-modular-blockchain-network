package server

import (
	"github.com/andantan/modular-blockchain/core/block"
	"github.com/andantan/modular-blockchain/network/message"
	"github.com/andantan/modular-blockchain/network/protocol/tcp"
	"github.com/andantan/modular-blockchain/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestServer_EncodeDecodeP2PMessage(t *testing.T) {
	s := &Server{}

	t.Run("transaction_encode_decode", func(t *testing.T) {
		tx := block.GenerateRandomTestTransaction(t)
		encoded, err := s.EncodeP2PMessage(tx)
		assert.NoError(t, err)
		assert.Equal(t, encoded[0], byte(message.MessageP2PType))
		assert.Equal(t, encoded[1], byte(message.MessageP2PSubTypeTransaction))

		addr := util.RandomAddress()
		rm := tcp.NewTcpRawMessage(addr, encoded)
		assert.Equal(t, "tcp", rm.Protocol())

		decoded, err := s.DecodeP2PMessage(rm)
		assert.NoError(t, err)
		assert.Equal(t, rm.Protocol(), decoded.Protocol)
		assert.True(t, rm.From().Equal(decoded.From))

		decodedTx, ok := decoded.Data.(*block.Transaction)
		assert.True(t, ok)

		txh, err := tx.Hash()
		assert.NoError(t, err)
		decodedTxh, err := decodedTx.Hash()
		assert.NoError(t, err)
		assert.True(t, txh.Equal(decodedTxh))
	})

	t.Run("block_encode_decode", func(t *testing.T) {
		b := block.GenerateRandomTestBlock(t, 1<<3)
		encoded, err := s.EncodeP2PMessage(b)
		assert.NoError(t, err)
		assert.Equal(t, encoded[0], byte(message.MessageP2PType))
		assert.Equal(t, encoded[1], byte(message.MessageP2PSubTypeBlock))

		addr := util.RandomAddress()
		rm := tcp.NewTcpRawMessage(addr, encoded)

		decoded, err := s.DecodeP2PMessage(rm)
		assert.NoError(t, err)
		assert.Equal(t, rm.Protocol(), decoded.Protocol)
		assert.True(t, rm.From().Equal(decoded.From))

		decodedB, ok := decoded.Data.(*block.Block)
		assert.True(t, ok)

		bh, err := b.Hash()
		assert.NoError(t, err)
		decodedBh, err := decodedB.Hash()
		assert.NoError(t, err)
		assert.True(t, bh.Equal(decodedBh))
	})

	t.Run("request_status_message_encode_decode", func(t *testing.T) {
		req := &message.RequestStatusMessage{}
		encoded, err := s.EncodeP2PMessage(req)
		assert.NoError(t, err)
		assert.Equal(t, encoded[0], byte(message.MessageP2PType))
		assert.Equal(t, encoded[1], byte(message.MessageP2PSubTypeRequestStatus))

		addr := util.RandomAddress()
		rm := tcp.NewTcpRawMessage(addr, encoded)

		decoded, err := s.DecodeP2PMessage(rm)
		assert.NoError(t, err)
		assert.Equal(t, rm.Protocol(), decoded.Protocol)
		assert.True(t, rm.From().Equal(decoded.From))

		_, ok := decoded.Data.(*message.RequestStatusMessage)
		assert.True(t, ok)
	})

	t.Run("response_status_message_encode_decode", func(t *testing.T) {
		addr := util.RandomAddress()
		res := &message.ResponseStatusMessage{
			Address:          addr,
			NetAddr:          "0.0.0.0:8080",
			Version:          uint32(2),
			Height:           uint64(55),
			State:            byte(Online),
			GenesisBlockHash: util.RandomHash(),
			CurrentBlockHash: util.RandomHash(),
		}

		encoded, err := s.EncodeP2PMessage(res)
		assert.NoError(t, err)
		assert.Equal(t, encoded[0], byte(message.MessageP2PType))
		assert.Equal(t, encoded[1], byte(message.MessageP2PSubTypeResponseStatus))

		rm := tcp.NewTcpRawMessage(addr, encoded)
		decoded, err := s.DecodeP2PMessage(rm)
		assert.NoError(t, err)
		assert.Equal(t, rm.Protocol(), decoded.Protocol)
		assert.True(t, rm.From().Equal(decoded.From))

		decodedRes, ok := decoded.Data.(*message.ResponseStatusMessage)
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
		req := &message.RequestHeadersMessage{
			From:  uint64(55),
			Count: uint64(11),
		}

		encoded, err := s.EncodeP2PMessage(req)
		assert.NoError(t, err)
		assert.Equal(t, encoded[0], byte(message.MessageP2PType))
		assert.Equal(t, encoded[1], byte(message.MessageP2PSubTypeRequestHeaders))

		rm := tcp.NewTcpRawMessage(util.RandomAddress(), encoded)
		decoded, err := s.DecodeP2PMessage(rm)
		assert.NoError(t, err)
		assert.Equal(t, rm.Protocol(), decoded.Protocol)
		assert.True(t, rm.From().Equal(decoded.From))

		decodedReq, ok := decoded.Data.(*message.RequestHeadersMessage)
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
		res := &message.ResponseHeadersMessage{Headers: headers}

		encoded, err := s.EncodeP2PMessage(res)
		assert.NoError(t, err)
		assert.Equal(t, encoded[0], byte(message.MessageP2PType))
		assert.Equal(t, encoded[1], byte(message.MessageP2PSubTypeResponseHeaders))

		rm := tcp.NewTcpRawMessage(util.RandomAddress(), encoded)
		decoded, err := s.DecodeP2PMessage(rm)
		assert.NoError(t, err)
		assert.Equal(t, rm.Protocol(), decoded.Protocol)
		assert.True(t, rm.From().Equal(decoded.From))

		decodedRes, ok := decoded.Data.(*message.ResponseHeadersMessage)
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
		req := &message.RequestBlocksMessage{
			From:  uint64(55),
			Count: uint64(11),
		}

		encoded, err := s.EncodeP2PMessage(req)
		assert.NoError(t, err)
		assert.Equal(t, encoded[0], byte(message.MessageP2PType))
		assert.Equal(t, encoded[1], byte(message.MessageP2PSubTypeRequestBlocks))

		rm := tcp.NewTcpRawMessage(util.RandomAddress(), encoded)
		decoded, err := s.DecodeP2PMessage(rm)
		assert.NoError(t, err)
		assert.Equal(t, rm.Protocol(), decoded.Protocol)
		assert.True(t, rm.From().Equal(decoded.From))

		decodedReq, ok := decoded.Data.(*message.RequestBlocksMessage)
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
		res := &message.ResponseBlocksMessage{Blocks: blocks}

		encoded, err := s.EncodeP2PMessage(res)
		assert.NoError(t, err)
		assert.Equal(t, encoded[0], byte(message.MessageP2PType))
		assert.Equal(t, encoded[1], byte(message.MessageP2PSubTypeResponseBlocks))

		rm := tcp.NewTcpRawMessage(util.RandomAddress(), encoded)
		decoded, err := s.DecodeP2PMessage(rm)
		assert.NoError(t, err)
		assert.Equal(t, rm.Protocol(), decoded.Protocol)
		assert.True(t, rm.From().Equal(decoded.From))

		decodedRes, ok := decoded.Data.(*message.ResponseBlocksMessage)
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
