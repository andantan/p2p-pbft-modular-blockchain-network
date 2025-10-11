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
