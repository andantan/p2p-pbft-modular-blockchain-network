package pbft

import (
	"github.com/andantan/modular-blockchain/core"
	"github.com/andantan/modular-blockchain/core/block"
	"github.com/andantan/modular-blockchain/crypto"
	"github.com/andantan/modular-blockchain/network/consensus"
	"github.com/andantan/modular-blockchain/types"
)

type PbftConsensusEngineFactory struct{}

func NewPbftConsensusEngineFactory() *PbftConsensusEngineFactory {
	return &PbftConsensusEngineFactory{}
}

func (f *PbftConsensusEngineFactory) NewConsensusEngine(
	k *crypto.PrivateKey,
	b *block.Block,
	p core.Processor,
	validators []types.Address,
) consensus.ConsensusEngine {
	return NewPbftConsensusEngine(k, b, p, validators)
}
