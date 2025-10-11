package protocol

import (
	"github.com/andantan/modular-blockchain/network/message"
	"github.com/andantan/modular-blockchain/types"
)

type Node interface {
	Listen()
	Connect(string) error
	ConsumeRawMessage() <-chan message.RawMessage
	Broadcast([]byte) error
	Peers() []types.Address
	Disconnect(types.Address)
	Close()
}
