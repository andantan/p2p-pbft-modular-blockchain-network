package network

import (
	"github.com/andantan/modular-blockchain/codec"
)

type Message interface {
	codec.Hasher     // for deterministic
	codec.Signer     // for verification
	codec.ProtoCodec // for encoding & decoding
}
