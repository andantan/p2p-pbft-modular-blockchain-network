package consensus

import (
	"github.com/andantan/modular-blockchain/core/block"
	"github.com/andantan/modular-blockchain/network/message"
	"github.com/andantan/modular-blockchain/types"
)

type ConsensusMessage interface {
	message.Message

	Address() types.Address
	Round() (uint64, uint64) // return view, sequence
	ProposalBlock() (*block.Block, bool)
}

type ConsensusMessageCodec interface {
	Encode(ConsensusMessage) ([]byte, error)
	Decode([]byte) (ConsensusMessage, error)
}
