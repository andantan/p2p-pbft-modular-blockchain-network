package pbft

import (
	"github.com/andantan/p2p-pbft-modular-blockchain-network/core"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/core/block"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPbftProposer_Createblock(t *testing.T) {
	bc := core.GenerateTestBlockchain(t)
	mp := core.GenerateTestMempool(t, 100)

	tx := block.GenerateRandomTestTransaction(t)
	assert.NoError(t, mp.Put(tx))

	proposer := GenerateTestPbftProposer(t)

	b, err := proposer.Createblock(bc, mp)
	assert.NoError(t, err)
	assert.NotNil(t, b)

	assert.Equal(t, uint64(1), b.Header.Height)
	assert.NoError(t, b.Verify())
	assert.Equal(t, uint64(1), b.Body.GetWeight())
}

func TestPbftProposer_ProposeBlock(t *testing.T) {
	proposer := GenerateTestPbftProposer(t)
	testBlock := block.GenerateRandomTestBlockWithHeight(t, 1<<4, 10)
	testBlock.Proposer = proposer.privKey.PublicKey()

	msg, err := proposer.ProposeBlock(testBlock)
	assert.NoError(t, err)

	prePrepareMsg, ok := msg.(*PbftPrePrepareMessage)
	assert.True(t, ok)

	assert.Equal(t, uint64(10), prePrepareMsg.Sequence)
	assert.NoError(t, prePrepareMsg.Verify())
}
