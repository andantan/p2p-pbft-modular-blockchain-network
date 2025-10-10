package codec

import "github.com/andantan/modular-blockchain/crypto"

type Signer interface {
	Sign(key *crypto.PrivateKey) error
	Verify() error
}
