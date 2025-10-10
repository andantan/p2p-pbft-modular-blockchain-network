package codec

import "github.com/andantan/modular-blockchain/types"

type Hasher interface {
	Hash() (types.Hash, error)
}
