package codec

import "github.com/andantan/p2p-pbft-modular-blockchain-network/types"

type Hasher interface {
	Hash() (types.Hash, error)
}
