package network

import "github.com/andantan/modular-blockchain/network/message"

type Broadcaster interface {
	Broadcast(message.Message) error
}
