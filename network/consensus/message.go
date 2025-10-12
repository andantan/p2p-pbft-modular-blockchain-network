package consensus

import (
	"github.com/andantan/modular-blockchain/network/message"
	"github.com/andantan/modular-blockchain/types"
)

type ConsensusMessage interface {
	message.Message

	Address() types.Address
}

type ConsensusMessageCodec interface {
	Encode(ConsensusMessage) ([]byte, error)
	Decode([]byte) (ConsensusMessage, error)
}
