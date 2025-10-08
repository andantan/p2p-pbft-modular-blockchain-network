package network

import "github.com/andantan/p2p-pbft-modular-blockchain-network/network/message"

type Broadcaster interface {
	Broadcast(message.Message) error
}
