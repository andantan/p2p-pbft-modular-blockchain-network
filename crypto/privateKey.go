package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
)

type PrivateKey struct {
	key *ecdsa.PrivateKey
}

func GeneratePrivateKey() (PrivateKey, error) {
	k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	if err != nil {
		return PrivateKey{}, err
	}

	return PrivateKey{
		k,
	}, nil
}

func (pk PrivateKey) PublicKey() PublicKey {
	k := pk.key.PublicKey

	return PublicKey{
		Key: elliptic.MarshalCompressed(k, k.X, k.Y),
	}
}

func (pk PrivateKey) Sign(data []byte) (Signature, error) {
	r, s, err := ecdsa.Sign(rand.Reader, pk.key, data)

	if err != nil {
		return Signature{}, err
	}

	return Signature{
		R: r,
		S: s,
	}, nil
}
