package message

import (
	"github.com/andantan/modular-blockchain/codec"
	"github.com/andantan/modular-blockchain/types"
)

type Raw interface {
	From() types.Address
	Payload() []byte
}

type Message interface {
	codec.Hasher     // for deterministic
	codec.Signer     // for verification
	codec.ProtoCodec // for encoding & decoding
}

type ConsensusMessage interface {
	Message

	Address() types.Address
}

type PeerInfo struct {
	Address        string `json:"address"`
	NetAddr        string `json:"net_addr"`
	Connections    uint8  `json:"connections"`
	MaxConnections uint8  `json:"max_connections"`
	Height         uint64 `json:"height"`
	IsValidator    bool   `json:"is_validator"`
}
