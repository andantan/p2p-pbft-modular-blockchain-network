package network

import "github.com/andantan/p2p-pbft-modular-blockchain-network/core/block"

type ConsensusEngine interface {
	StartEngine()
	HandleMessage(ConsensusMessage)
	OutgoingMessage() <-chan ConsensusMessage
	FinalizedBlock() <-chan *block.Block
	StopEngine()
}
