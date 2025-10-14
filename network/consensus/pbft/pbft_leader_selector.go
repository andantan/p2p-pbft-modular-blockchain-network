package pbft

import (
	"github.com/andantan/modular-blockchain/types"
	"sort"
)

type PbftLeaderSelector struct{}

func NewPbftLeaderSelector() *PbftLeaderSelector {
	return &PbftLeaderSelector{}
}

func (s *PbftLeaderSelector) SelectLeader(validators []types.Address, round uint64) types.Address {
	sort.Slice(validators, func(i, j int) bool {
		return validators[i].Lte(validators[j])
	})

	leaderIndex := int(round % uint64(len(validators)))

	return validators[leaderIndex]
}
