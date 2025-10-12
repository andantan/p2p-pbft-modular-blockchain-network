package synchronizer

import (
	"errors"
	"fmt"
	"github.com/andantan/modular-blockchain/codec"
	"github.com/andantan/modular-blockchain/network/message"
)

const (
	MessageSyncSubTypeRequestStatus message.MessageSyncSubType = iota
	MessageSyncSubTypeResponseStatus
	MessageSyncSubTypeRequestHeaders
	MessageSyncSubTypeResponseHeaders
	MessageSyncSubTypeRequestBlocks
	MessageSyncSubTypeResponseBlocks
)

type SyncMessageCodec interface {
	Encode(SyncMessage) ([]byte, error)
	Decode([]byte) (SyncMessage, error)
}

type DefaultSyncMessageCodec struct{}

func NewDefaultSyncMessageCodec() *DefaultSyncMessageCodec {
	return &DefaultSyncMessageCodec{}
}

func (c *DefaultSyncMessageCodec) Encode(msg SyncMessage) ([]byte, error) {
	msgType := message.MessageSyncType

	var subType message.MessageSyncSubType
	switch t := msg.(type) {
	case *RequestStatusMessage:
		subType = MessageSyncSubTypeRequestStatus
	case *ResponseStatusMessage:
		subType = MessageSyncSubTypeResponseStatus
	case *RequestHeadersMessage:
		subType = MessageSyncSubTypeRequestHeaders
	case *ResponseHeadersMessage:
		subType = MessageSyncSubTypeResponseHeaders
	case *RequestBlocksMessage:
		subType = MessageSyncSubTypeRequestBlocks
	case *ResponseBlocksMessage:
		subType = MessageSyncSubTypeResponseBlocks
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

func (c *DefaultSyncMessageCodec) Decode(payload []byte) (SyncMessage, error) {
	if len(payload) < 2 {
		return nil, errors.New("payload too short")
	}

	msgType := message.MessageType(payload[0])
	if msgType != message.MessageSyncType {
		return nil, fmt.Errorf("invalid message type: expected Sync, got %d", msgType)
	}

	subType := message.MessageSyncSubType(payload[1])
	data := payload[2:]

	var o codec.ProtoCodec
	switch subType {
	case MessageSyncSubTypeRequestStatus:
		o = new(RequestStatusMessage)
	case MessageSyncSubTypeResponseStatus:
		o = new(ResponseStatusMessage)
	case MessageSyncSubTypeRequestHeaders:
		o = new(RequestHeadersMessage)
	case MessageSyncSubTypeResponseHeaders:
		o = new(ResponseHeadersMessage)
	case MessageSyncSubTypeRequestBlocks:
		o = new(RequestBlocksMessage)
	case MessageSyncSubTypeResponseBlocks:
		o = new(ResponseBlocksMessage)
	default:
		return nil, fmt.Errorf("invalid message type [%d, %d]", msgType, subType)
	}

	if err := codec.DecodeProto(data, o); err != nil {
		return nil, err
	}

	return o, nil
}
