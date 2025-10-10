package network

import (
	"github.com/andantan/modular-blockchain/core/block"
	"github.com/andantan/modular-blockchain/network/message"
)

type ConsensusEngine interface {
	StartEngine()
	HandleMessage(message.ConsensusMessage)
	OutgoingMessage() <-chan message.ConsensusMessage
	FinalizedBlock() <-chan *block.Block
	StopEngine()
}
