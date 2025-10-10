package network

import (
	"github.com/andantan/modular-blockchain/network/message"
	"github.com/andantan/modular-blockchain/types"
)

type Node interface {
	Listen()
	Connect(string) error
	ConsumeRawMessage() <-chan message.Raw
	Broadcast(message.Message) error
	Disconnect(types.Address)
	Close()
}
