package message

import (
	"github.com/andantan/modular-blockchain/codec"
	"github.com/andantan/modular-blockchain/types"
)

type Message interface {
	codec.Hasher     // for deterministic
	codec.Signer     // for verification
	codec.ProtoCodec // for encoding & decoding
}

type RawMessage interface {
	Protocol() string
	From() types.Address
	Payload() []byte
}

type DecodedMessage struct {
	Protocol string
	From     types.Address
	Data     any
}

type MessageType byte

const (
	MessageP2PType MessageType = iota
	MessageConsensusType
)

type MessageP2PSubType byte

const (
	MessageP2PSubTypeTransaction MessageP2PSubType = iota
	MessageP2PSubTypeBlock
	MessageP2PSubTypeRequestStatus
	MessageP2PSubTypeResponseStatus
	MessageP2PSubTypeRequestHeaders
	MessageP2PSubTypeResponseHeaders
	MessageP2PSubTypeRequestBlocks
	MessageP2PSubTypeResponseBlocks
)

type MessageConsensusSubType byte
