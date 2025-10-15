package pbft

import (
	"github.com/andantan/modular-blockchain/core/block"
	"github.com/andantan/modular-blockchain/network/consensus"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestPbftConsensusEngine_HandlePrePrepare(t *testing.T) {
	vc, ih := 4, 5
	_, validatorKeys, engine, _, msgCh := GenerateTestPbftConsensusEngine(t, vc, ih)

	leaderKey := GetLeaderFromTestValidators(t, validatorKeys, 0, uint64(ih+1))
	b := engine.block
	ppMsg := NewPbftPrePrepareMessage(0, b.Header.Height, b, leaderKey.PublicKey())
	assert.NoError(t, ppMsg.Sign(leaderKey))

	engine.Start()

	go engine.HandleMessage(ppMsg)

	select {
	case outgoingMsg := <-msgCh:
		_, ok := outgoingMsg.(*PbftPrePrepareMessage)
		assert.True(t, ok)
		assert.True(t, engine.state.Eq(PrePrepared))

	case <-time.After(100 * time.Millisecond):
		t.Fatal("engine does not send PrePrepareMessage to channel")
	}

	select {
	case outgoingMsg := <-msgCh:
		_, ok := outgoingMsg.(*PbftPrepareMessage)
		assert.True(t, ok)
		assert.True(t, engine.state.Eq(PrePrepared))

	case <-time.After(100 * time.Millisecond):
		t.Fatal("engine does not send PrePrepareMessage to channel")
	}

	engine.Stop()
}

func TestPbftConsensusEngine_HandlePrepare(t *testing.T) {
	// phase pre-prepare
	vc, ih := 10, 6
	_, validatorKeys, engine, _, msgCh := GenerateTestPbftConsensusEngine(t, vc, ih)
	leaderKey := GetLeaderFromTestValidators(t, validatorKeys, 0, uint64(ih))

	b := engine.block
	ppMsg := NewPbftPrePrepareMessage(0, uint64(ih), b, leaderKey.PublicKey())
	assert.NoError(t, ppMsg.Sign(leaderKey))

	engine.Start()

	var wg sync.WaitGroup
	handler := func(msg consensus.ConsensusMessage) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			engine.HandleMessage(msg)
		}()
	}

	handler(ppMsg)

	// gossip test
	select {
	case outgoingMsg := <-msgCh:
		_, ok := outgoingMsg.(*PbftPrePrepareMessage)
		assert.True(t, ok)
		assert.True(t, engine.state.Eq(PrePrepared))

	case <-time.After(100 * time.Millisecond):
		t.Fatal("engine does not send PrePrepareMessage to channel")
	}

	// state change test
	select {
	case outgoingMsg := <-msgCh:
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

		handler(pm)
	}

	// gossip test
	for i := 0; i < engine.quorum; i++ {
		select {
		case outgoingMsg := <-msgCh:
			_, ok := outgoingMsg.(*PbftPrepareMessage)
			assert.True(t, ok)

		case <-time.After(100 * time.Millisecond):
			t.Fatal("engine does not send PrePrepareMessage to channel")
		}
	}

	select {
	case outgoingMsg := <-msgCh:
		_, ok := outgoingMsg.(*PbftCommitMessage)
		assert.True(t, ok)
		assert.True(t, engine.state.Eq(Prepared))

	case <-time.After(100 * time.Millisecond):
		t.Fatal("engine does not send PrePrepareMessage to channel")
	}

	wg.Wait()

	engine.Stop()
}

func TestPbftConsensusEngine_HandleCommit(t *testing.T) {
	// phase pre-prepare
	vc, ih := 15, 6
	nh := ih + 1
	_, validatorKeys, engine, fbCh, msgCh := GenerateTestPbftConsensusEngine(t, vc, ih)
	leaderKey := GetLeaderFromTestValidators(t, validatorKeys, 0, uint64(nh))

	b := engine.block
	height := b.Header.Height
	ppMsg := NewPbftPrePrepareMessage(0, height, b, leaderKey.PublicKey())
	assert.NoError(t, ppMsg.Sign(leaderKey))

	engine.Start()

	var wg sync.WaitGroup
	handler := func(msg consensus.ConsensusMessage) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			engine.HandleMessage(msg)
		}()
	}

	handler(ppMsg)

	select {
	case outgoingMsg := <-msgCh:
		_, ok := outgoingMsg.(*PbftPrePrepareMessage)
		assert.True(t, ok)

	case <-time.After(100 * time.Millisecond):
		t.Fatal("engine does not send PrePrepareMessage to channel")
	}

	select {
	case outgoingMsg := <-msgCh:
		_, ok := outgoingMsg.(*PbftPrepareMessage)
		assert.True(t, ok)

	case <-time.After(100 * time.Millisecond):
		t.Fatal("engine does not send PrePrepareMessage to channel!")
	}

	// phase prepare

	bh, err := b.Hash()
	assert.NoError(t, err)

	for i := 0; i < engine.quorum; i++ {
		pm := NewPbftPrepareMessage(0, height, bh)
		assert.NoError(t, pm.Sign(validatorKeys[i]))

		handler(pm)
	}

	for i := 0; i < engine.quorum; i++ {
		select {
		case outgoingMsg := <-msgCh:
			_, ok := outgoingMsg.(*PbftPrepareMessage)
			assert.True(t, ok)

		case <-time.After(100 * time.Millisecond):
			t.Fatal("engine does not send PrePrepareMessage to channel")
		}
	}

	select {
	case outgoingMsg := <-msgCh:
		_, ok := outgoingMsg.(*PbftCommitMessage)
		assert.True(t, ok)

	case <-time.After(100 * time.Millisecond):
		t.Fatal("engine does not send PrePrepareMessage to channel")
	}

	for i := 0; i < engine.quorum; i++ {
		pm := NewPbftCommitMessage(0, height, bh)
		assert.NoError(t, pm.Sign(validatorKeys[i]))

		handler(pm)
	}

	for i := 0; i < engine.quorum-1; i++ {
		select {
		case outgoingMsg := <-msgCh:
			_, ok := outgoingMsg.(*PbftCommitMessage)
			assert.True(t, ok)

		case <-time.After(100 * time.Millisecond):
			t.Fatal("engine does not send PrePrepareMessage to channel")
		}
	}

	select {
	case outgoingMsg := <-msgCh:
		_, ok := outgoingMsg.(*PbftCommitMessage)
		assert.True(t, ok)

	case <-time.After(100 * time.Millisecond):
		t.Fatal("engine does not send PrePrepareMessage to channel")
	}

	select {
	case fb := <-fbCh:
		finalizedBlockHash, err := fb.Hash()
		assert.True(t, engine.state.Eq(Finalized))
		assert.NoError(t, err)
		assert.True(t, bh.Equal(finalizedBlockHash))
		assert.NotNil(t, fb.Tail)

	case <-time.After(time.Second):
		t.Fatal("block does not finalized")
	}

	wg.Wait()

	engine.Stop()
	assert.True(t, engine.state.Eq(Terminated))
}

func TestPbftConsensusEngine_ViewChange(t *testing.T) {
	vc, ih := 15, 6
	keys, engines := GenerateTestPbftMultipleConsensusEngine(t, vc, ih)

	t.Cleanup(func() {
		for _, engine := range engines {
			engine.Stop()
		}
	})

	quorum := (2 * len(keys) / 3) + 1
	sequence := engines[0].sequence
	for _, engine := range engines {
		assert.Equal(t, quorum, engine.quorum)
		assert.Equal(t, sequence, engine.sequence)
	}

	igMsgCh := make(chan consensus.ConsensusMessage, 100)
	igFbCh := make(chan *block.Block, 100)

	for _, engine := range engines {
		go func(e *PbftConsensusEngine) {
			e.Start()

			msgCh := e.OutgoingMessage()
			fbCh := e.FinalizedBlock()

			for {
				timer := time.After(20 * time.Second)

				select {
				case <-e.closeCh:
					return
				case <-timer:
					addr := e.validator.PublicKey().Address()
					t.Errorf("engine does not send FinalizedMessage to channel (%s)\n", addr.ShortString(8))
					return
				case msg := <-msgCh:
					igMsgCh <- msg
				case fb := <-fbCh:
					igFbCh <- fb
					return
				}
			}
		}(engine)
	}

	var wg sync.WaitGroup
	handler := func(engine *PbftConsensusEngine, msg consensus.ConsensusMessage) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			engine.HandleMessage(msg)
		}()
	}

	newView := uint64(1)
	for i := 0; i < vc; i++ {
		vcMsg := NewPbftViewChangeMessage(newView, sequence)
		assert.NoError(t, vcMsg.Sign(keys[i]))

		for _, engine := range engines {
			handler(engine, vcMsg)
		}
	}

	finalizedCount := 0
	timer := time.After(20 * time.Second)
	for finalizedCount < vc {
		select {
		case <-timer:
			t.Fatalf("time-out (%d/%d)\n", finalizedCount, vc)
		case msg := <-igMsgCh:
			for _, engine := range engines {
				handler(engine, msg)
			}
		case <-igFbCh:
			finalizedCount++
		}
	}

	wg.Wait()

	<-time.After(500 * time.Millisecond)

	assert.Equal(t, finalizedCount, vc)

	for _, e := range engines {
		assert.True(t, e.state.Eq(Finalized))
	}
}
