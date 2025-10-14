package consensus

import (
	"github.com/andantan/modular-blockchain/core"
	"github.com/andantan/modular-blockchain/core/block"
)

type Proposer interface {
	Createblock(core.Chain, core.VirtualMemoryPool) (*block.Block, error)
	ProposeBlock(*block.Block) (ConsensusMessage, error)
}
