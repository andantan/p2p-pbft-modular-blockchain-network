package protocol

import (
	"github.com/andantan/modular-blockchain/network"
	"github.com/andantan/modular-blockchain/types"
)

type Node interface {
	Listen()
	Connect(string) error
	ConsumeRawMessage() <-chan Raw
	Broadcast(network.Message) error
	Disconnect(types.Address)
	Close()
}
