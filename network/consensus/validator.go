package consensus

import (
	"github.com/andantan/modular-blockchain/crypto"
	"github.com/andantan/modular-blockchain/types"
)

type Validator interface {
	PublicKey() *crypto.PublicKey
	Sign(ConsensusMessage) error
	Verify(ConsensusMessage) error
	UpdateValidatorSet([]types.Address)
	ProcessConsensusMessage(ConsensusMessage) error
}
