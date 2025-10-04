package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

const (
	SignatureLength = 64
)

type Signature struct {
	R *big.Int
	S *big.Int
}

func (s *Signature) IsNil() bool {
	return s.R == nil || s.S == nil
}

func (s *Signature) Verify(k *PublicKey, d []byte) bool {
	x, y := elliptic.UnmarshalCompressed(elliptic.P256(), k.Key)
	key := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}

	return ecdsa.Verify(key, d, s.R, s.S)
}

func (s *Signature) Bytes() []byte {
	if s.IsNil() {
		return nil
	}

	rBytes := make([]byte, 32)
	sBytes := make([]byte, 32)

	s.R.FillBytes(rBytes)
	s.S.FillBytes(sBytes)

	return append(rBytes, sBytes...)
}

func (s *Signature) String() string {
	b := s.Bytes()

	if b == nil {
		return ""
	}

	return hex.EncodeToString(b)
}

func SignatureFromBytes(b []byte) (*Signature, error) {
	if len(b) != SignatureLength {
		return nil, fmt.Errorf("invalid signature length: %d", len(b))
	}

	r := new(big.Int).SetBytes(b[:32])
	s := new(big.Int).SetBytes(b[32:])

	return &Signature{
		R: r,
		S: s,
	}, nil
}

func SignatureFromHexString(s string) (*Signature, error) {
	s = strings.TrimPrefix(s, "0x")

	if len(s) != SignatureLength*2 {
		return nil, fmt.Errorf("invalid hex string length (%d), must be %d", len(s), SignatureLength*2)
	}

	sigBytes, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}

	return SignatureFromBytes(sigBytes)
}
