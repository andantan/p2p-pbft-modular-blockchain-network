package network

import (
	"github.com/andantan/modular-blockchain/crypto"
	"github.com/andantan/modular-blockchain/network/message"
	"github.com/andantan/modular-blockchain/types"
)

type Validator interface {
	PublicKey() *crypto.PublicKey
	Sign(message.ConsensusMessage) error
	Verify(message.ConsensusMessage) error
	UpdateValidatorSet([]types.Address)
	ProcessConsensusMessage(message.ConsensusMessage) error
}
