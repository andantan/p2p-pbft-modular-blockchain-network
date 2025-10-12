package synchronizer

import (
	"github.com/andantan/modular-blockchain/core"
	"github.com/andantan/modular-blockchain/core/block"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestChainSynchronizer_Synchronize(t *testing.T) {
	bc1, s1 := GenerateTestSynchronizer(t)
	bc2, s2 := GenerateTestSynchronizer(t)

	targetHeight := 10

	core.AddTestBlocksToBlockchain(t, bc2, uint64(targetHeight))

	s1.Start()
	s2.Start()

	t.Cleanup(func() {
		s1.Stop()
		assert.True(t, s1.state.Eq(Terminated))
		s2.Stop()
		assert.True(t, s2.state.Eq(Terminated))
	})

	b2, err := bc2.GetCurrentBlock()
	assert.NoError(t, err)
	h2, err := b2.Hash()
	assert.NoError(t, err)

	status2 := &ResponseStatusMessage{
		Address:          s2.address,
		NetAddr:          s2.netAddr,
		Version:          b2.Header.Version,
		Height:           bc2.GetCurrentHeight(),
		GenesisBlockHash: s2.genesisBlockHash,
		CurrentBlockHash: h2,
	}

	s1.HandleMessage(s2.address, status2)
	time.Sleep(100 * time.Millisecond)
	s1.synchronize()

	var s1ReqBlocksMsg *RequestBlocksMessage
	select {
	case msg := <-s1.OutgoingMessage():
		directedMsg, ok := msg.(*DirectedMessage)
		assert.True(t, ok)
		assert.True(t, directedMsg.To.Equal(s2.address))
		s1ReqBlocksMsg, ok = directedMsg.Message.(*RequestBlocksMessage)
		assert.True(t, ok)
		assert.Equal(t, s1ReqBlocksMsg.From, uint64(1))
		assert.Equal(t, s1ReqBlocksMsg.Count, uint64(16))
		assert.True(t, s1.state.Eq(SyncingBlocks))
	case <-time.After(100 * time.Millisecond):
		t.Fatal("time-out")
	}

	s2.HandleMessage(s1.address, s1ReqBlocksMsg)
	time.Sleep(100 * time.Millisecond)

	var s2ResBlocksMsg *ResponseBlocksMessage
	select {
	case msg := <-s2.OutgoingMessage():
		directedMsg, ok := msg.(*DirectedMessage)
		assert.True(t, ok)
		assert.True(t, directedMsg.To.Equal(s1.address))
		s2ResBlocksMsg, ok = directedMsg.Message.(*ResponseBlocksMessage)
		assert.True(t, ok)
		assert.Equal(t, targetHeight, len(s2ResBlocksMsg.Blocks))
	case <-time.After(100 * time.Millisecond):
		t.Fatal("time-out")
	}

	s1.HandleMessage(s2.address, s2ResBlocksMsg)

	ticker := time.NewTicker(50 * time.Millisecond)
	for {
		<-ticker.C
		if bc1.GetCurrentHeight() == uint64(10) {
			break
		}
	}
	ticker.Stop()

	assert.Equal(t, bc1.GetCurrentHeight(), bc2.GetCurrentHeight())

	for i := uint64(0); i <= uint64(targetHeight); i++ {
		b1, err := bc1.GetBlockByHeight(i)
		assert.NoError(t, err)
		b2, err := bc2.GetBlockByHeight(i)
		assert.NoError(t, err)
		h1, err := b1.Hash()
		assert.NoError(t, err)
		h2, err := b2.Hash()
		assert.NoError(t, err)
		assert.True(t, h1.Equal(h2))
	}

	assert.True(t, s1.state.Eq(Idle))
	assert.True(t, s2.state.Eq(Idle))

	afterB1, err := bc1.GetCurrentBlock()
	assert.NoError(t, err)
	afterH1, err := afterB1.Hash()
	assert.NoError(t, err)

	afterStatus1 := &ResponseStatusMessage{
		Address:          s1.address,
		NetAddr:          s1.netAddr,
		Version:          afterB1.Header.Version,
		Height:           bc1.GetCurrentHeight(),
		GenesisBlockHash: s1.genesisBlockHash,
		CurrentBlockHash: afterH1,
	}

	afterB2, err := bc2.GetCurrentBlock()
	assert.NoError(t, err)
	afterH2, err := afterB2.Hash()
	assert.NoError(t, err)

	afterStatus2 := &ResponseStatusMessage{
		Address:          s2.address,
		NetAddr:          s2.netAddr,
		Version:          afterB2.Header.Version,
		Height:           bc2.GetCurrentHeight(),
		GenesisBlockHash: s2.genesisBlockHash,
		CurrentBlockHash: afterH2,
	}

	s2.HandleMessage(s1.address, afterStatus1)
	s1.HandleMessage(s2.address, afterStatus2)

	time.Sleep(100 * time.Millisecond)

	s1.synchronize()
	s2.synchronize()

	time.Sleep(50 * time.Millisecond)

	assert.True(t, s1.state.Eq(Synchronized))
	assert.True(t, s2.state.Eq(Synchronized))
}

func TestChainSynchronizer_ForkResolution(t *testing.T) {
	bc1, s1 := GenerateTestSynchronizer(t)
	bc2, s2 := GenerateTestSynchronizer(t)

	genesisHeader, err := bc1.GetCurrentHeader()
	assert.NoError(t, err)

	height1Block := block.GenerateRandomTestBlockWithPrevHeader(t, genesisHeader, 1<<4)

	assert.NoError(t, bc1.AddBlock(height1Block)) // height 1
	assert.NoError(t, bc2.AddBlock(height1Block)) // height 1

	// forked height = 2
	adds := 3
	core.AddTestBlocksToBlockchain(t, bc1, uint64(2)) // height 3
	assert.Equal(t, uint64(3), bc1.GetCurrentHeight())
	core.AddTestBlocksToBlockchain(t, bc2, uint64(adds)) // height 4
	assert.Equal(t, uint64(4), bc2.GetCurrentHeight())

	s1.Start()
	s2.Start()

	t.Cleanup(func() {
		s1.Stop()
		assert.True(t, s1.state.Eq(Terminated))
		s2.Stop()
		assert.True(t, s2.state.Eq(Terminated))
	})

	b2, err := bc2.GetCurrentBlock()
	assert.NoError(t, err)
	h2, err := b2.Hash()
	assert.NoError(t, err)

	status2 := &ResponseStatusMessage{
		Address:          s2.address,
		NetAddr:          s2.netAddr,
		Version:          b2.Header.Version,
		Height:           bc2.GetCurrentHeight(),
		GenesisBlockHash: s2.genesisBlockHash,
		CurrentBlockHash: h2,
	}

	s1.HandleMessage(s2.address, status2)
	time.Sleep(100 * time.Millisecond)
	s1.synchronize()

	var s1ReqBlocksMsg *RequestBlocksMessage
	select {
	case msg := <-s1.OutgoingMessage():
		directedMsg, ok := msg.(*DirectedMessage)
		assert.True(t, ok)
		assert.True(t, directedMsg.To.Equal(s2.address))
		s1ReqBlocksMsg, ok = directedMsg.Message.(*RequestBlocksMessage)
		assert.True(t, ok)
		assert.Equal(t, s1ReqBlocksMsg.From, uint64(4))
		assert.Equal(t, s1ReqBlocksMsg.Count, uint64(16))
		assert.True(t, s1.state.Eq(SyncingBlocks))
	case <-time.After(100 * time.Millisecond):
		t.Fatal("time-out")
	}

	s2.HandleMessage(s1.address, s1ReqBlocksMsg)
	time.Sleep(100 * time.Millisecond)

	var s2ResBlocksMsg *ResponseBlocksMessage
	select {
	case msg := <-s2.OutgoingMessage():
		directedMsg, ok := msg.(*DirectedMessage)
		assert.True(t, ok)
		assert.True(t, directedMsg.To.Equal(s1.address))
		s2ResBlocksMsg, ok = directedMsg.Message.(*ResponseBlocksMessage)
		assert.True(t, ok)
		assert.Equal(t, 1, len(s2ResBlocksMsg.Blocks))
	case <-time.After(100 * time.Millisecond):
		t.Fatal("time-out")
	}

	s1.HandleMessage(s2.address, s2ResBlocksMsg)
	time.Sleep(200 * time.Millisecond)

	assert.True(t, s1.state.Eq(ResolvingFork))

	var s1ReqHeadersMsg *RequestHeadersMessage
	select {
	case msg := <-s1.OutgoingMessage():
		directedMsg, ok := msg.(*DirectedMessage)
		assert.True(t, ok)
		assert.True(t, directedMsg.To.Equal(s2.address))
		s1ReqHeadersMsg, ok = directedMsg.Message.(*RequestHeadersMessage)
		assert.True(t, ok)
		assert.Equal(t, s1ReqHeadersMsg.From, uint64(3))
		assert.Equal(t, s1ReqHeadersMsg.Count, uint64(100))
	case <-time.After(100 * time.Millisecond):
		t.Fatal("time-out")
	}

	s2.HandleMessage(s1.address, s1ReqHeadersMsg)
	time.Sleep(100 * time.Millisecond)

	var s2ResHeadersMsg *ResponseHeadersMessage
	select {
	case msg := <-s2.OutgoingMessage():
		directedMsg, ok := msg.(*DirectedMessage)
		assert.True(t, ok)
		assert.True(t, directedMsg.To.Equal(s1.address))
		s2ResHeadersMsg, ok = directedMsg.Message.(*ResponseHeadersMessage)
		assert.True(t, ok)
		assert.Equal(t, 2, len(s2ResHeadersMsg.Headers))
	case <-time.After(100 * time.Millisecond):
		t.Fatal("time-out")
	}

	s1.HandleMessage(s2.address, s2ResHeadersMsg)
	time.Sleep(100 * time.Millisecond)

	assert.True(t, s1.state.Eq(SyncingBlocks))
	assert.True(t, s2.state.Eq(Idle))
	assert.Equal(t, bc1.GetCurrentHeight(), uint64(1))
	assert.Equal(t, bc2.GetCurrentHeight(), uint64(4))

	var afterS1ReqBlocksMsg *RequestBlocksMessage
	select {
	case msg := <-s1.OutgoingMessage():
		directedMsg, ok := msg.(*DirectedMessage)
		assert.True(t, ok)
		assert.True(t, directedMsg.To.Equal(s2.address))
		afterS1ReqBlocksMsg, ok = directedMsg.Message.(*RequestBlocksMessage)
		assert.True(t, ok)
		assert.Equal(t, uint64(2), afterS1ReqBlocksMsg.From)
		assert.Equal(t, uint64(16), afterS1ReqBlocksMsg.Count)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("time-out")
	}

	s2.HandleMessage(s1.address, afterS1ReqBlocksMsg)
	time.Sleep(100 * time.Millisecond)

	var afterS2ResBlocksMsg *ResponseBlocksMessage
	select {
	case msg := <-s2.OutgoingMessage():
		directedMsg, ok := msg.(*DirectedMessage)
		assert.True(t, ok)
		assert.True(t, directedMsg.To.Equal(s1.address))
		afterS2ResBlocksMsg, ok = directedMsg.Message.(*ResponseBlocksMessage)
		assert.True(t, ok)
		assert.Equal(t, 3, len(afterS2ResBlocksMsg.Blocks))
	case <-time.After(100 * time.Millisecond):
		t.Fatal("time-out")
	}

	s1.HandleMessage(s2.address, afterS2ResBlocksMsg)
	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, bc1.GetCurrentHeight(), bc2.GetCurrentHeight())
	assert.True(t, s1.state.Eq(Idle))
	assert.True(t, s2.state.Eq(Idle))

	afterB1, err := bc1.GetCurrentBlock()
	assert.NoError(t, err)
	afterH1, err := afterB1.Hash()
	assert.NoError(t, err)

	afterStatus1 := &ResponseStatusMessage{
		Address:          s1.address,
		NetAddr:          s1.netAddr,
		Version:          afterB1.Header.Version,
		Height:           bc1.GetCurrentHeight(),
		GenesisBlockHash: s1.genesisBlockHash,
		CurrentBlockHash: afterH1,
	}

	afterB2, err := bc2.GetCurrentBlock()
	assert.NoError(t, err)
	afterH2, err := afterB2.Hash()
	assert.NoError(t, err)

	afterStatus2 := &ResponseStatusMessage{
		Address:          s2.address,
		NetAddr:          s2.netAddr,
		Version:          afterB2.Header.Version,
		Height:           bc2.GetCurrentHeight(),
		GenesisBlockHash: s2.genesisBlockHash,
		CurrentBlockHash: afterH2,
	}

	s2.HandleMessage(s1.address, afterStatus1)
	s1.HandleMessage(s2.address, afterStatus2)

	time.Sleep(100 * time.Millisecond)

	s1.synchronize()
	s2.synchronize()

	time.Sleep(50 * time.Millisecond)

	assert.True(t, s1.state.Eq(Synchronized))
	assert.True(t, s2.state.Eq(Synchronized))
}
