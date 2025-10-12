package message

import (
	"errors"
	"fmt"
	"github.com/andantan/modular-blockchain/codec"
	"github.com/andantan/modular-blockchain/core/block"
)

// GossipMessage for block, transaction
type GossipMessage interface {
	codec.ProtoCodec
}

const (
	MessageGossipSubTypeTransaction MessageGossipSubType = iota
	MessageGossipSubTypeBlock
)

type GossipMessageCodec interface {
	Encode(GossipMessage) ([]byte, error)
	Decode([]byte) (GossipMessage, error)
}

type DefaultGossipMessageCodec struct{}

func NewDefaultGossipMessageCodec() *DefaultGossipMessageCodec {
	return &DefaultGossipMessageCodec{}
}

func (*DefaultGossipMessageCodec) Encode(msg GossipMessage) ([]byte, error) {
	msgType := MessageGossipType

	var subType MessageGossipSubType
	switch t := msg.(type) {
	case *block.Transaction:
		subType = MessageGossipSubTypeTransaction
	case *block.Block:
		subType = MessageGossipSubTypeBlock
	default:
		return nil, fmt.Errorf("unknown p2p message type: %T", t)
	}

	data, err := codec.EncodeProto(msg)
	if err != nil {
		return nil, err
	}

	payload := make([]byte, 0)
	payload = append(payload, byte(msgType))
	payload = append(payload, byte(subType))
	payload = append(payload, data...)

	return payload, nil
}

func (*DefaultGossipMessageCodec) Decode(payload []byte) (GossipMessage, error) {
	if len(payload) < 2 {
		return nil, errors.New("payload too short")
	}

	msgType := MessageType(payload[0])
	if msgType != MessageGossipType {
		return nil, fmt.Errorf("invalid message type: expected Sync, got %d", msgType)
	}

	subType := MessageGossipSubType(payload[1])
	data := payload[2:]

	var o codec.ProtoCodec
	switch subType {
	case MessageGossipSubTypeTransaction:
		o = new(block.Transaction)
	case MessageGossipSubTypeBlock:
		o = new(block.Block)
	default:
		return nil, fmt.Errorf("invalid message type [%d, %d]", msgType, subType)
	}

	if err := codec.DecodeProto(data, o); err != nil {
		return nil, err
	}

	return o, nil
}
