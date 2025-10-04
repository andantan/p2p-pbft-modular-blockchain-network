package crypto

import (
	"crypto/sha256"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/types"
)

type PublicKey struct {
	Key []byte
}

func (pk PublicKey) Address() types.Address {
	h := sha256.Sum256(pk.Key)
	s := len(h) - types.AddressLength
	a, _ := types.AddressFromBytes(h[s:])

	return a
}
