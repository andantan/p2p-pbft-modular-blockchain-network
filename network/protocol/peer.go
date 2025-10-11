package protocol

import (
	"github.com/andantan/modular-blockchain/crypto"
	"github.com/andantan/modular-blockchain/types"
)

type Peer interface {
	Send([]byte) error
	Read()
	Close()
	PublicKey() *crypto.PublicKey
	Address() types.Address
	NetAddr() string
	ConsumeRawMessage() <-chan Raw
}
