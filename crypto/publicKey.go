package crypto

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/types"
)

const (
	// PublicKeyLength Standard length of a P256 compressed public key
	PublicKeyLength = 33
)

type PublicKey struct {
	Key []byte
}

func (k *PublicKey) Bytes() []byte {
	b := make([]byte, len(k.Key))

	copy(b, k.Key)

	return b
}

func PublicKeyFromBytes(b []byte) (*PublicKey, error) {
	if len(b) != PublicKeyLength {
		return nil, fmt.Errorf("invalid public key length: %d, must be %d", len(b), PublicKeyLength)
	}

	k := &PublicKey{
		Key: make([]byte, PublicKeyLength),
	}

	copy(k.Key, b)

	return k, nil
}

func (k *PublicKey) Address() types.Address {
	h := sha256.Sum256(k.Key)
	s := len(h) - types.AddressLength
	a, _ := types.AddressFromBytes(h[s:])

	return a
}

func (k *PublicKey) String() string {
	if k == nil || len(k.Key) == 0 {
		return "PublicKey<nil>"
	}

	return fmt.Sprintf("PublicKey{%s...}", hex.EncodeToString(k.Key)[:16])
}

func (k *PublicKey) Equal(other *PublicKey) bool {
	return bytes.Equal(k.Key, other.Key)
}
