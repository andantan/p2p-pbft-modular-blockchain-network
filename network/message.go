package network

import (
	"github.com/andantan/p2p-pbft-modular-blockchain-network/types"
	"io"
)

type RawMessage struct {
	From    types.Address
	Payload io.Reader
}
