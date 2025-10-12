package consensus

import (
	"github.com/andantan/modular-blockchain/core/block"
)

type ConsensusEngine interface {
	Start()
	HandleMessage(ConsensusMessage)
	OutgoingMessage() <-chan ConsensusMessage
	FinalizedBlock() <-chan *block.Block
	Stop()
}
