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
	Payload() []byte // [type, sub-type, data...]
}

type MessageType byte

const (
	MessageGossipType MessageType = iota
	MessageSyncType
	MessageConsensusType
)

type MessageGossipSubType byte
type MessageSyncSubType byte
type MessageConsensusSubType byte
