package consensus

import (
	"github.com/andantan/modular-blockchain/types"
)

type LeaderSelector interface {
	SelectLeader([]types.Address, uint64) types.Address
}
