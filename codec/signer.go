package codec

import "github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"

type Signer interface {
	Sign(key *crypto.PrivateKey) error
	Verify() error
}
