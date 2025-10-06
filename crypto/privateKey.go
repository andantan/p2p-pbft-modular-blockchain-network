package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
)

type PrivateKey struct {
	key *ecdsa.PrivateKey
}

func GeneratePrivateKey() (*PrivateKey, error) {
	k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	if err != nil {
		return nil, err
	}

	return &PrivateKey{
		k,
	}, nil
}

func (k *PrivateKey) PublicKey() *PublicKey {
	pk := k.key.PublicKey

	return &PublicKey{
		Key: elliptic.MarshalCompressed(pk, pk.X, pk.Y),
	}
}

func (k *PrivateKey) Sign(data []byte) (*Signature, error) {
	r, s, err := ecdsa.Sign(rand.Reader, k.key, data)

	if err != nil {
		return nil, err
	}

	return &Signature{
		R: r,
		S: s,
	}, nil
}

func (k *PrivateKey) String() string {
	if k == nil || k.key == nil {
		return "PrivateKey<nil>"
	}

	return fmt.Sprintf("PrivateKey{D: %s...}", k.key.D.Text(16)[:8])
}
