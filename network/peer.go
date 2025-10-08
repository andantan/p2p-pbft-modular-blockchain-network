package network

import (
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/types"
)

type Peer interface {
	PublicKey() *crypto.PublicKey
	Address() types.Address
	NetAddr() string
	Handshake(Message) (Message, error)
	Send([]byte) error
	Read()
	Close()
}
