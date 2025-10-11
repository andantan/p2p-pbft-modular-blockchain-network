package consensus

import (
	"github.com/andantan/modular-blockchain/network"
	"github.com/andantan/modular-blockchain/types"
)

type ConsensusMessage interface {
	network.Message

	Address() types.Address
}
