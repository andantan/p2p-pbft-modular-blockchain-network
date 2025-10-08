package network

import (
	"bytes"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/core/block"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/network/message"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/types"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/util"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func TestPbftValidator_Verify(t *testing.T) {
	valN := 5
	validator, keys := GenerateTestPbftValidator(t, valN)

	for i := 0; i < valN; i++ {
		msg := message.NewPbftPrepareMessage(0, 0, util.RandomHash())
		assert.NoError(t, msg.Sign(keys[i]))
		assert.NoError(t, validator.Verify(msg))
	}

	nonValidatorKey, _ := crypto.GeneratePrivateKey()
	msgInvalid := message.NewPbftPrepareMessage(0, 0, util.RandomHash())
	assert.NoError(t, msgInvalid.Sign(nonValidatorKey))
	assert.Error(t, validator.Verify(msgInvalid))
}

func TestPbftValidator_GetLeader(t *testing.T) {
	validator, _ := GenerateTestPbftValidator(t, 4)
	sortedAddrs := make([]types.Address, 4)
	copy(sortedAddrs, validator.validatorSet.Values())
	sort.Slice(sortedAddrs, func(i, j int) bool {
		return bytes.Compare(sortedAddrs[i].Bytes(), sortedAddrs[j].Bytes()) < 0
	})

	// scenario 1: sequence(height) 1, view 0
	// (1 + 0) % 4 = 1
	expectedLeader1 := sortedAddrs[1]
	actualLeader1 := validator.getLeader(0, 1)
	assert.True(t, expectedLeader1.Equal(actualLeader1))

	// scenario 2: sequence(height) 5, view 0
	// (5 + 0) % 4 = 1
	expectedLeader2 := sortedAddrs[1]
	actualLeader2 := validator.getLeader(0, 5)
	assert.True(t, expectedLeader2.Equal(actualLeader2))

	// scenario 3: sequence(height) 1, view 1
	// (1 + 1) % 4 = 2
	expectedLeader3 := sortedAddrs[2]
	actualLeader3 := validator.getLeader(1, 1)
	assert.True(t, expectedLeader3.Equal(actualLeader3))
}

func TestPbftValidator_ProcessPrePrepare_LeaderValidation(t *testing.T) {
	validator, keys := GenerateTestPbftValidator(t, 4)
	addrs := make([]types.Address, 4)
	for i, k := range keys {
		addrs[i] = k.PublicKey().Address()
	}

	height := uint64(4)
	view := uint64(0)

	sort.Slice(addrs, func(i, j int) bool {
		return bytes.Compare(addrs[i].Bytes(), addrs[j].Bytes()) < 0
	})

	leaderIndex := int((height + view) % uint64(len(addrs)))
	expectedLeaderAddr := addrs[leaderIndex]

	var leaderKey *crypto.PrivateKey
	for _, k := range keys {
		if k.PublicKey().Address().Equal(expectedLeaderAddr) {
			leaderKey = k
			break
		}
	}
	assert.NotNil(t, leaderKey)

	// Success scenario
	t.Run("it_should_process_message_from_valid_leader", func(t *testing.T) {
		b := block.GenerateRandomTestBlockWithHeight(t, 10, height)
		assert.NoError(t, b.Sign(leaderKey))

		msg := message.NewPbftPrePrepareMessage(view, height, b, leaderKey.PublicKey())
		assert.NoError(t, msg.Sign(leaderKey))

		assert.NoError(t, validator.ProcessConsensusMessage(msg))
	})

	// failure scenario
	t.Run("it_should_reject_message_from_invalid_leader", func(t *testing.T) {
		var invalidLeaderKey *crypto.PrivateKey
		for _, k := range keys {
			if !k.PublicKey().Address().Equal(expectedLeaderAddr) {
				invalidLeaderKey = k
				break
			}
		}

		b := block.GenerateRandomTestBlockWithHeight(t, 10, height)

		msg := message.NewPbftPrePrepareMessage(view, height, b, invalidLeaderKey.PublicKey())
		assert.NoError(t, msg.Sign(invalidLeaderKey))

		err := validator.ProcessConsensusMessage(msg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid leader")
	})
}
