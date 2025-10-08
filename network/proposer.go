package network

import (
	"github.com/andantan/p2p-pbft-modular-blockchain-network/core"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/core/block"
)

type Proposer interface {
	Createblock(core.Chain, core.VirtualMemoryPool) (*block.Block, error)
	ProposeBlock(*block.Block) (Message, error)
}
