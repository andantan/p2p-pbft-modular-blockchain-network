package consensus

import (
	"github.com/andantan/modular-blockchain/core/block"
)

type ConsensusEngine interface {
	StartEngine()
	HandleMessage(ConsensusMessage)
	OutgoingMessage() <-chan ConsensusMessage
	FinalizedBlock() <-chan *block.Block
	StopEngine()
}
