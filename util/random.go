package util

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"github.com/andantan/modular-blockchain/types"
)

func RandomBytes(len int) []byte {
	b := make([]byte, len)
	_, _ = rand.Read(b)

	return b
}

func RandomInt() int {
	b := RandomBytes(8)

	return int(binary.LittleEndian.Uint64(b))
}

func RandomIntWithMaximun(m int) int {
	return RandomInt() % m
}

func RandomUint64() uint64 {
	b := RandomBytes(8)

	return binary.LittleEndian.Uint64(b)
}

func RandomUint64WithMaximun(m int) uint64 {
	r := RandomUint64()
	if m < 0 {
		return r
	}
	return r % uint64(m)
}

func RandomHash() types.Hash {
	return sha256.Sum256(RandomBytes(types.HashLength))
}
