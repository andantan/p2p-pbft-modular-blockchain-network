package types

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	AddressLength = 20
)

type Address [AddressLength]uint8

func (a Address) Bytes() []byte {
	return a[:]
}

func (a Address) IsZero() bool {
	return a == Address{}
}

func (a Address) String() string {
	return "0x" + hex.EncodeToString(a[:])
}

func (a Address) ShortString(l int) string {
	as := hex.EncodeToString(a[:])

	if l > len(as) {
		l = len(as)
	}

	return "0x" + as[:l]
}

func (a Address) Equal(other Address) bool {
	return bytes.Equal(a[:], other[:])
}

func (a Address) Gt(other Address) bool {
	return bytes.Compare(a[:], other[:]) > 0
}

func (a Address) Gte(other Address) bool {
	return bytes.Compare(a[:], other[:]) >= 0
}

func (a Address) Lt(other Address) bool {
	return bytes.Compare(a[:], other[:]) < 0
}

func (a Address) Lte(other Address) bool {
	return bytes.Compare(a[:], other[:]) <= 0
}

func AddressFromBytes(b []byte) (Address, error) {
	if len(b) != AddressLength {
		return Address{}, fmt.Errorf("given bytes with address-length %d should be 20 bytes", len(b))
	}

	var a Address

	copy(a[:], b)

	return a, nil
}

func AddressFromHexString(s string) (Address, error) {
	s = strings.TrimPrefix(s, "0x")

	if len(s) != AddressLength*2 {
		return Address{}, fmt.Errorf("invalid hex string length (%d), must be %d", len(s), AddressLength*2)
	}

	b, err := hex.DecodeString(s)
	if err != nil {
		return Address{}, err
	}

	return AddressFromBytes(b)
}
