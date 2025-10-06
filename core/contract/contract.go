package contract

import "github.com/andantan/p2p-pbft-modular-blockchain-network/core/block"

type Contract interface {
	Execute(transaction *block.Transaction) error
}
