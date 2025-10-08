package network

import (
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/types"
)

type Validator interface {
	PublicKey() *crypto.PublicKey
	Sign(ConsensusMessage) error
	Verify(ConsensusMessage) error
	UpdateValidatorSet([]types.Address)
	ProcessConsensusMessage(ConsensusMessage) error
}
