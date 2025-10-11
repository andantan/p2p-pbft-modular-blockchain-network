package message

import (
	"fmt"
	"github.com/andantan/modular-blockchain/core/block"
	pbSync "github.com/andantan/modular-blockchain/proto/network/message"
	"github.com/andantan/modular-blockchain/types"
	"google.golang.org/protobuf/proto"
)

type RequestStatusMessage struct{}

func (m *RequestStatusMessage) ToProto() (proto.Message, error) {
	return &pbSync.RequestStatusMessage{}, nil
}

func (m *RequestStatusMessage) FromProto(msg proto.Message) error {
	_, ok := msg.(*pbSync.RequestStatusMessage)
	if !ok {
		return fmt.Errorf("invalid proto message type for RequestStatusMessage")
	}

	return nil
}

func (m *RequestStatusMessage) EmptyProto() proto.Message {
	return &pbSync.RequestStatusMessage{}
}

type ResponseStatusMessage struct {
	Address          types.Address
	NetAddr          string
	Version          uint32
	Height           uint64
	State            byte
	GenesisBlockHash types.Hash
	CurrentBlockHash types.Hash
}

func (m *ResponseStatusMessage) ToProto() (proto.Message, error) {
	return &pbSync.ResponseStatusMessage{
		Address:          m.Address.Bytes(),
		NetAddr:          m.NetAddr,
		Version:          m.Version,
		Height:           m.Height,
		State:            uint32(m.State),
		GenesisBlockHash: m.GenesisBlockHash.Bytes(),
		CurrentBlockHash: m.CurrentBlockHash.Bytes(),
	}, nil
}

func (m *ResponseStatusMessage) FromProto(msg proto.Message) error {
	p, ok := msg.(*pbSync.ResponseStatusMessage)
	if !ok {
		return fmt.Errorf("invalid proto message type for ResponseStatusMessage")
	}

	var (
		err  error
		addr types.Address
		gbh  types.Hash
		cbh  types.Hash
	)

	if addr, err = types.AddressFromBytes(p.Address); err != nil {
		return err
	}

	if gbh, err = types.HashFromBytes(p.GenesisBlockHash); err != nil {
		return err
	}

	if cbh, err = types.HashFromBytes(p.CurrentBlockHash); err != nil {
		return err
	}

	m.Address = addr
	m.NetAddr = p.NetAddr
	m.Version = p.Version
	m.Height = p.Height
	m.State = byte(p.State)
	m.GenesisBlockHash = gbh
	m.CurrentBlockHash = cbh

	return nil
}

func (m *ResponseStatusMessage) EmptyProto() proto.Message {
	return &pbSync.ResponseStatusMessage{}
}

type RequestHeadersMessage struct {
	From  uint64
	Count uint64
}

func (m *RequestHeadersMessage) ToProto() (proto.Message, error) {
	return &pbSync.RequestHeadersMessage{
		From:  m.From,
		Count: m.Count,
	}, nil
}

func (m *RequestHeadersMessage) FromProto(msg proto.Message) error {
	p, ok := msg.(*pbSync.RequestHeadersMessage)

	if !ok {
		return fmt.Errorf("invalid proto message type for RequestHeadersMessage")
	}

	m.From = p.From
	m.Count = p.Count

	return nil
}

func (m *RequestHeadersMessage) EmptyProto() proto.Message {
	return &pbSync.RequestHeadersMessage{}
}

type ResponseHeadersMessage struct {
	Headers []*block.Header
}

func (m *ResponseHeadersMessage) ToProto() (proto.Message, error) {
	headers, err := block.HeadersToProto(m.Headers)
	if err != nil {
		return nil, err
	}

	return &pbSync.ResponseHeadersMessage{
		Headers: headers,
	}, nil
}

func (m *ResponseHeadersMessage) FromProto(msg proto.Message) error {
	p, ok := msg.(*pbSync.ResponseHeadersMessage)
	if !ok {
		return fmt.Errorf("invalid proto message type for ResponseHeadersMessage")
	}

	var (
		err     error
		headers []*block.Header
	)

	if headers, err = block.HeadersFromProto(p.Headers); err != nil {
		return err
	}

	m.Headers = headers

	return nil
}

func (m *ResponseHeadersMessage) EmptyProto() proto.Message {
	return &pbSync.ResponseHeadersMessage{}
}

type RequestBlocksMessage struct {
	From  uint64
	Count uint64
}

func (m *RequestBlocksMessage) ToProto() (proto.Message, error) {
	return &pbSync.RequestBlocksMessage{
		From:  m.From,
		Count: m.Count,
	}, nil
}

func (m *RequestBlocksMessage) FromProto(msg proto.Message) error {
	p, ok := msg.(*pbSync.RequestBlocksMessage)

	if !ok {
		return fmt.Errorf("invalid proto message type for RequestBlocksMessage")
	}

	m.From = p.From
	m.Count = p.Count

	return nil
}

func (m *RequestBlocksMessage) EmptyProto() proto.Message {
	return &pbSync.RequestBlocksMessage{}
}

type ResponseBlocksMessage struct {
	Blocks []*block.Block
}

func (m *ResponseBlocksMessage) ToProto() (proto.Message, error) {
	blocks, err := block.BlocksToProto(m.Blocks)
	if err != nil {
		return nil, err
	}

	return &pbSync.ResponseBlocksMessage{
		Blocks: blocks,
	}, nil
}

func (m *ResponseBlocksMessage) FromProto(msg proto.Message) error {
	p, ok := msg.(*pbSync.ResponseBlocksMessage)
	if !ok {
		return fmt.Errorf("invalid proto message type for ResponseBlocksMessage")
	}

	var (
		err    error
		blocks []*block.Block
	)

	if blocks, err = block.BlocksFromProto(p.Blocks); err != nil {
		return err
	}

	m.Blocks = blocks

	return nil
}

func (m *ResponseBlocksMessage) EmptyProto() proto.Message {
	return &pbSync.ResponseBlocksMessage{}
}
