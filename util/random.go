package util

import (
	"crypto/rand"
	"crypto/sha256"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/types"
)

func RandomBytes(len int) []byte {
	b := make([]byte, len)
	_, _ = rand.Read(b)

	return b
}

func RandomHash() types.Hash {
	return sha256.Sum256(RandomBytes(types.HashLength))
}
