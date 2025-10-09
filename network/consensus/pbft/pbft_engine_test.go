package pbft

import (
	"github.com/andantan/p2p-pbft-modular-blockchain-network/core/block"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestPbftConsensusEngine_HandlePrePrepare(t *testing.T) {
	vc, ih := 4, 5
	bc, validatorKeys, engine, _, extCh := GenerateTestPbftConsensusEngine(t, vc, ih)

	leaderKey := GetLeaderFromTestValidators(t, validatorKeys, 0, uint64(ih+1))

	ph, err := bc.GetCurrentHeader()
	assert.NoError(t, err)

	b := block.GenerateRandomTestBlockWithPrevHeaderAndKey(t, ph, 1<<3, leaderKey)
	ppMsg := NewPbftPrePrepareMessage(0, b.Header.Height, b, leaderKey.PublicKey())
	assert.NoError(t, ppMsg.Sign(leaderKey))

	engine.StartEngine()

	go engine.HandleMessage(ppMsg)

	select {
	case outgoingMsg := <-extCh:
		_, ok := outgoingMsg.(*PbftPrepareMessage)
		assert.True(t, ok)
		assert.True(t, engine.state.Eq(PrePrepared))

	case <-time.After(100 * time.Millisecond):
		t.Fatal("engine does not send PrePrepareMessage to channel")
	}

	engine.StopEngine()
}

func TestPbftConsensusEngine_HandlePrepare(t *testing.T) {
	// phase pre-prepare
	vc, ih := 10, 6
	bc, validatorKeys, engine, _, extCh := GenerateTestPbftConsensusEngine(t, vc, ih)
	leaderKey := GetLeaderFromTestValidators(t, validatorKeys, 0, uint64(ih))
	ph, err := bc.GetCurrentHeader()
	assert.NoError(t, err)

	b := block.GenerateRandomTestBlockWithPrevHeaderAndKey(t, ph, 1<<3, leaderKey)
	ppMsg := NewPbftPrePrepareMessage(0, uint64(ih), b, leaderKey.PublicKey())
	assert.NoError(t, ppMsg.Sign(leaderKey))

	engine.StartEngine()

	go engine.HandleMessage(ppMsg)

	select {
	case outgoingMsg := <-extCh:
		_, ok := outgoingMsg.(*PbftPrepareMessage)
		assert.True(t, ok)
		assert.True(t, engine.state.Eq(PrePrepared))

	case <-time.After(100 * time.Millisecond):
		t.Fatal("engine does not send PrePrepareMessage to channel")
	}

	// phase prepare

	bh, err := b.Hash()
	assert.NoError(t, err)

	for i := 0; i < engine.quorum; i++ {
		pm := NewPbftPrepareMessage(0, uint64(ih), bh)
		assert.NoError(t, pm.Sign(validatorKeys[i]))

		go engine.HandleMessage(pm)

		if i != engine.quorum-1 {
			select {
			case outgoingMsg := <-extCh:
				_, ok := outgoingMsg.(*PbftPrepareMessage)
				assert.True(t, ok)
				assert.True(t, engine.state.Eq(PrePrepared))
			case <-time.After(100 * time.Millisecond):
				t.Fatal("engine does not send PrePrepareMessage to channel")
			}
		} else {
			select {
			case outgoingMsg := <-extCh:
				_, ok := outgoingMsg.(*PbftCommitMessage)
				assert.True(t, ok)
				assert.True(t, engine.state.Eq(Prepared))

			case <-time.After(100 * time.Millisecond):
				t.Fatal("engine does not send PrePrepareMessage to channel")
			}
		}
	}

	engine.StopEngine()
}

func TestPbftConsensusEngine_HandleCommit(t *testing.T) {
	// phase pre-prepare
	vc, ih := 15, 6
	nh := ih + 1
	bc, validatorKeys, engine, fbCh, extCh := GenerateTestPbftConsensusEngine(t, vc, ih)
	leaderKey := GetLeaderFromTestValidators(t, validatorKeys, 0, uint64(nh))
	ph, err := bc.GetCurrentHeader()
	assert.NoError(t, err)

	b := block.GenerateRandomTestBlockWithPrevHeaderAndKey(t, ph, 1<<3, leaderKey)
	height := b.Header.Height
	ppMsg := NewPbftPrePrepareMessage(0, height, b, leaderKey.PublicKey())
	assert.NoError(t, ppMsg.Sign(leaderKey))

	engine.StartEngine()

	go engine.HandleMessage(ppMsg)

	select {
	case outgoingMsg := <-extCh:
		_, ok := outgoingMsg.(*PbftPrepareMessage)
		assert.True(t, ok)
		assert.True(t, engine.state.Eq(PrePrepared))

	case <-time.After(100 * time.Millisecond):
		t.Fatal("engine does not send PrePrepareMessage to channel!")
	}

	// phase prepare

	bh, err := b.Hash()
	assert.NoError(t, err)

	for i := 0; i < engine.quorum; i++ {
		pm := NewPbftPrepareMessage(0, height, bh)
		assert.NoError(t, pm.Sign(validatorKeys[i]))

		go engine.HandleMessage(pm)

		if i != engine.quorum-1 {
			select {
			case outgoingMsg := <-extCh:
				_, ok := outgoingMsg.(*PbftPrepareMessage)
				assert.True(t, ok)
				assert.True(t, engine.state.Eq(PrePrepared))
			case <-time.After(100 * time.Millisecond):
				t.Fatal("engine does not send PrePrepareMessage to channel")
			}
		} else {
			select {
			case outgoingMsg := <-extCh:
				_, ok := outgoingMsg.(*PbftCommitMessage)
				assert.True(t, ok)
				assert.True(t, engine.state.Eq(Prepared))

			case <-time.After(100 * time.Millisecond):
				t.Fatal("engine does not send PrePrepareMessage to channel")
			}
		}
	}

	for i := 0; i < engine.quorum; i++ {
		pm := NewPbftCommitMessage(0, height, bh)
		assert.NoError(t, pm.Sign(validatorKeys[i]))

		go engine.HandleMessage(pm)

		if i != engine.quorum-1 {
			select {
			case outgoingMsg := <-extCh:
				_, ok := outgoingMsg.(*PbftCommitMessage)
				assert.True(t, ok)
				assert.True(t, engine.state.Eq(Prepared))
			case <-time.After(100 * time.Millisecond):
				t.Fatal("engine does not send PbftCommitMessage to channel!")
			}
		}
	}

	select {
	case fb := <-fbCh:
		finalizedBlockHash, err := fb.Hash()
		assert.NoError(t, err)
		assert.True(t, bh.Equal(finalizedBlockHash))
		assert.NotNil(t, fb.Tail)

	case <-time.After(time.Second):
		t.Fatal("block does not finalized")
	}

	engine.StopEngine()
	assert.True(t, engine.state.Eq(Terminated))
}
