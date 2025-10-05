package types

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	HashLength = 32
)

type Hash [HashLength]uint8

func (h Hash) Bytes() []byte {
	return h[:]
}

func (h Hash) IsZero() bool {
	return h == Hash{}
}

func (h Hash) String() string {
	return "0x" + hex.EncodeToString(h[:])
}

func (h Hash) ShortString(l int) string {
	hs := hex.EncodeToString(h[:])

	if l > len(hs) {
		l = len(hs)
	}

	return "0x" + hs[:l]
}

func (h Hash) Eq(o Hash) bool {
	return h == o
}

func (h Hash) Gt(other Hash) bool {
	return bytes.Compare(h[:], other[:]) > 0
}

func (h Hash) Gte(other Hash) bool {
	return bytes.Compare(h[:], other[:]) >= 0
}

func (h Hash) Lt(other Hash) bool {
	return bytes.Compare(h[:], other[:]) < 0
}

func (h Hash) Lte(other Hash) bool {
	return bytes.Compare(h[:], other[:]) <= 0
}

func HashFromBytes(b []byte) (Hash, error) {
	if len(b) != HashLength {
		return Hash{}, fmt.Errorf("given bytes with hash-length %d should be 32 bytes", len(b))
	}

	var h Hash

	copy(h[:], b)

	return h, nil
}

func HashFromHexString(s string) (Hash, error) {
	s = strings.TrimPrefix(s, "0x")

	if len(s) != HashLength*2 {
		return Hash{}, fmt.Errorf("invalid hex string length (%d), must be %d", len(s), HashLength*2)
	}

	b, err := hex.DecodeString(s)
	if err != nil {
		return Hash{}, err
	}

	return HashFromBytes(b)
}

func FilledHash(b byte) Hash {
	h := Hash{}

	for i := range HashLength {
		h[i] = b
	}

	return h
}
