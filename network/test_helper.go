package network

import (
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func GenerateTestPbftValidator(t *testing.T, valN int) (*PbftValidator, []*crypto.PrivateKey) {
	t.Helper()

	keys := make([]*crypto.PrivateKey, valN)
	addrs := make([]types.Address, valN)

	for i := 0; i < valN; i++ {
		keys[i], _ = crypto.GeneratePrivateKey()
		addrs[i] = keys[i].PublicKey().Address()
		assert.NotNil(t, keys[i])
		assert.NotNil(t, addrs[i])
	}

	validator := NewPbftValidator(keys[0])
	validator.UpdateValidatorSet(addrs)

	return validator, keys
}

func GenerateTestPbftProposer(t *testing.T) *PbftProposer {
	t.Helper()

	k, _ := crypto.GenerateTestKeyPair(t)
	p := NewPbftProposer(k)

	assert.NotNil(t, p)

	return p
}
