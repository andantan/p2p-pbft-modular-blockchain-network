package network

import "github.com/andantan/p2p-pbft-modular-blockchain-network/types"

type Node interface {
	Listen()
	Connect(string) error
	Remove(types.Address)
	Stop()
}
