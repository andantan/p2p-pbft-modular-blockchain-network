package consensus

import (
	"github.com/andantan/modular-blockchain/core"
	"github.com/andantan/modular-blockchain/core/block"
	"github.com/andantan/modular-blockchain/crypto"
	"github.com/andantan/modular-blockchain/types"
)

type ConsensusEngineFactory interface {
	NewConsensusEngine(*crypto.PrivateKey, *block.Block, core.Processor, []types.Address) ConsensusEngine
}
