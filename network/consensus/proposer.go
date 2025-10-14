package consensus

import (
	"github.com/andantan/modular-blockchain/core/block"
)

type Proposer interface {
	Createblock() (*block.Block, error)
	ProposeBlock(*block.Block) (ConsensusMessage, error)
}
