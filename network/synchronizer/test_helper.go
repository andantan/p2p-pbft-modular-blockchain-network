package synchronizer

import (
	"github.com/andantan/modular-blockchain/core"
	"github.com/andantan/modular-blockchain/util"
	"testing"
)

func GenerateTestSynchronizer(t *testing.T) (*core.Blockchain, *ChainSynchroizer) {
	bc := core.GenerateTestBlockchain(t)
	s := NewChainSynchronizer(util.RandomAddress(), "our-net-addr", bc)

	return bc, s
}
