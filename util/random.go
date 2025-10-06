package util

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/types"
)

func RandomBytes(len int) []byte {
	b := make([]byte, len)
	_, _ = rand.Read(b)

	return b
}

func RandomUint64() uint64 {
	b := RandomBytes(8)

	return binary.LittleEndian.Uint64(b)
}

func RandomHash() types.Hash {
	return sha256.Sum256(RandomBytes(types.HashLength))
}
