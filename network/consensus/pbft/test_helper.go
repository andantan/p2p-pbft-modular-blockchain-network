package pbft

import (
	"bytes"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/core/block"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/types"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func GenerateTestPbftPrePrepareMessage(t *testing.T, view uint64) (*PbftPrePrepareMessage, *crypto.PrivateKey) {
	t.Helper()

	privKey, _ := crypto.GeneratePrivateKey()
	b := block.GenerateRandomTestBlock(t, 1<<4)
	msg := NewPbftPrePrepareMessage(view, b.Header.Height, b, privKey.PublicKey())

	assert.NoError(t, msg.Sign(privKey))
	assert.NotNil(t, msg.PublicKey)
	assert.NotNil(t, msg.Signature)
	assert.NoError(t, msg.Verify())

	return msg, privKey
}

func GenerateTestPbftPrePrepareMessageWithSequence(t *testing.T, view, sequence uint64) (*PbftPrePrepareMessage, *crypto.PrivateKey) {
	t.Helper()

	privKey, _ := crypto.GeneratePrivateKey()
	b := block.GenerateRandomTestBlockWithHeight(t, 10, sequence)
	msg := NewPbftPrePrepareMessage(view, sequence, b, privKey.PublicKey())

	assert.NoError(t, msg.Sign(privKey))
	assert.NotNil(t, msg.PublicKey)
	assert.NotNil(t, msg.Signature)
	assert.NoError(t, msg.Verify())

	return msg, privKey
}

func GenerateTestPbftPrepareMessage(t *testing.T, view uint64) (*PbftPrepareMessage, *crypto.PrivateKey) {
	t.Helper()

	privKey, _ := crypto.GeneratePrivateKey()
	b := block.GenerateRandomTestBlock(t, 1<<4)
	h, err := b.Hash()
	assert.NoError(t, err)
	msg := NewPbftPrepareMessage(view, b.Header.Height, h)

	assert.NoError(t, msg.Sign(privKey))
	assert.NotNil(t, msg.PublicKey)
	assert.NotNil(t, msg.Signature)
	assert.NoError(t, msg.Verify())

	return msg, privKey
}

func GenerateTestPbftCommitMessage(t *testing.T, view uint64) (*PbftCommitMessage, *crypto.PrivateKey) {
	t.Helper()

	privKey, _ := crypto.GeneratePrivateKey()
	b := block.GenerateRandomTestBlock(t, 1<<4)
	h, err := b.Hash()
	assert.NoError(t, err)
	msg := NewPbftCommitMessage(view, b.Header.Height, h)

	assert.NoError(t, msg.Sign(privKey))
	assert.NotNil(t, msg.PublicKey)
	assert.NotNil(t, msg.Signature)
	assert.NoError(t, msg.Verify())

	return msg, privKey
}

func GenerateTestPbftViewChangeMessage(t *testing.T, view, sequence uint64) (*PbftViewChangeMessage, *crypto.PrivateKey) {
	t.Helper()

	privKey, _ := crypto.GeneratePrivateKey()
	msg := NewPbftViewChangeMessage(view, sequence)
	assert.NoError(t, msg.Sign(privKey))
	assert.NotNil(t, msg.PublicKey)
	assert.NotNil(t, msg.Signature)
	assert.NoError(t, msg.Verify())

	return msg, privKey
}

func GenerateTestPbftNewViewMessageWithKey(t *testing.T, view, sequence uint64, vcmN int, k *crypto.PrivateKey) (*PbftNewViewMessage, *crypto.PrivateKey) {
	t.Helper()

	prePrepareMsg, _ := GenerateTestPbftPrePrepareMessageWithSequence(t, view, sequence)

	vcms := make([]*PbftViewChangeMessage, vcmN)

	for i := 0; i < vcmN; i++ {
		vcm, _ := GenerateTestPbftViewChangeMessage(t, view, sequence)
		vcms[i] = vcm
	}

	msg := &PbftNewViewMessage{
		NewView:            view,
		Sequence:           sequence,
		ViewChangeMessages: vcms,
		PrePrepareMessage:  prePrepareMsg,
	}
	assert.NoError(t, msg.Sign(k))
	assert.NotNil(t, msg.PublicKey)
	assert.NotNil(t, msg.Signature)
	assert.NoError(t, msg.Verify())

	return msg, k
}

func GenerateTestPbftNewViewMessage(t *testing.T, view, sequence uint64, vcmN int) (*PbftNewViewMessage, *crypto.PrivateKey) {
	t.Helper()

	privKey, _ := crypto.GeneratePrivateKey()

	return GenerateTestPbftNewViewMessageWithKey(t, view, sequence, vcmN, privKey)
}

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

func GetLeaderFromTestValidators(t *testing.T, vs []*crypto.PrivateKey, view, sequence uint64) *crypto.PrivateKey {
	t.Helper()

	addrs := make([]types.Address, 4)
	for i, k := range vs {
		addrs[i] = k.PublicKey().Address()
	}

	sort.Slice(addrs, func(i, j int) bool {
		return bytes.Compare(addrs[i].Bytes(), addrs[j].Bytes()) < 0
	})

	leaderIndex := int((view + sequence) % uint64(len(addrs)))
	expectedLeaderAddr := addrs[leaderIndex]

	var leaderKey *crypto.PrivateKey
	for _, k := range vs {
		if k.PublicKey().Address().Equal(expectedLeaderAddr) {
			leaderKey = k
			break
		}
	}
	assert.NotNil(t, leaderKey)

	return leaderKey
}

func GenerateTestPbftProposer(t *testing.T) *PbftProposer {
	t.Helper()

	k, _ := crypto.GenerateTestKeyPair(t)
	p := NewPbftProposer(k)

	assert.NotNil(t, p)

	return p
}
