package server

import (
	"fmt"
	"github.com/andantan/modular-blockchain/core"
	"github.com/andantan/modular-blockchain/core/block"
	"github.com/andantan/modular-blockchain/crypto"
	"github.com/andantan/modular-blockchain/network/message"
	"github.com/andantan/modular-blockchain/network/protocol/tcp"
	"github.com/andantan/modular-blockchain/network/synchronizer"
	"github.com/andantan/modular-blockchain/types"
	"github.com/andantan/modular-blockchain/util"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestServer_ProcessGossipMessage(t *testing.T) {
	mainNodeAddr := ":10600"
	maxPeers := 7
	mainNode := tcp.GenerateTestTcpNode(t, mainNodeAddr, maxPeers)

	mainNode.Listen()
	time.Sleep(100 * time.Millisecond)

	numPeers := 7
	remoteNodes := make([]*tcp.TcpNode, numPeers)
	for i := 0; i < numPeers; i++ {
		time.Sleep(50 * time.Millisecond)
		go func() {
			remoteNode := tcp.GenerateTestTcpNode(t, fmt.Sprintf(":%d", 10500+i), maxPeers)
			remoteNode.Listen()

			time.Sleep(100 * time.Millisecond)
			remoteNodes[i] = remoteNode
			assert.NoError(t, remoteNode.Connect(mainNodeAddr))
		}()
	}

	t.Cleanup(func() {
		mainNode.Close()
		for _, n := range remoteNodes {
			n.Close()
		}
	})

	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, maxPeers, len(mainNode.Peers()))
	for i := 0; i < numPeers; i++ {
		assert.Equal(t, 1, len(remoteNodes[i].Peers()))
	}

	bc := core.GenerateTestBlockchain(t)
	mp := core.GenerateTestMempool(t, 100)
	c := message.NewDefaultGossipMessageCodec()
	s := &Server{
		logger: util.LoggerWithPrefixes("Server"),
		state:  types.NewAtomicNumber[ServerState](Online),
	}
	s.Node = mainNode
	s.Chain = bc
	s.VirtualMemoryPool = mp
	s.GossipMessageCodec = c

	tx := block.GenerateRandomTestTransaction(t)
	txHash, err := tx.Hash()
	assert.NoError(t, err)
	txPayload, err := c.Encode(tx)
	assert.NoError(t, err)
	txRm := tcp.NewTcpRawMessage(util.RandomAddress(), txPayload)
	s.processRawMessage(txRm)

	time.Sleep(300 * time.Millisecond)

	wgTx := new(sync.WaitGroup)
	wgTx.Add(numPeers)

	for i, remoteNode := range remoteNodes {
		go func() {
			defer wgTx.Done()
			select {
			case msg := <-remoteNode.ConsumeRawMessage():
				assert.NotEmpty(t, msg.Payload())
				decodedMsg, err := c.Decode(msg.Payload())
				assert.NoError(t, err)
				decodedTx, ok := decodedMsg.(*block.Transaction)
				assert.True(t, ok)
				h, err := decodedTx.Hash()
				assert.NoError(t, err)
				assert.True(t, h.Equal(txHash))
			case <-time.After(500 * time.Millisecond):
				t.Errorf("peer %d does not received msg", i)
			}
		}()
	}

	wgTx.Wait()

	genesis, err := bc.GetHeaderByHeight(0)
	assert.NoError(t, err)
	b := block.GenerateRandomTestBlockWithPrevHeader(t, genesis, 1<<10)
	blockHash, err := b.Hash()
	assert.NoError(t, err)
	blockPayload, err := c.Encode(b)
	assert.NoError(t, err)
	blockRm := tcp.NewTcpRawMessage(util.RandomAddress(), blockPayload)
	s.processRawMessage(blockRm)

	time.Sleep(300 * time.Millisecond)

	wgBlock := new(sync.WaitGroup)
	wgBlock.Add(numPeers)

	for i, remoteNode := range remoteNodes {
		go func() {
			defer wgBlock.Done()
			select {
			case msg := <-remoteNode.ConsumeRawMessage():
				assert.NotEmpty(t, msg.Payload())
				decodedMsg, err := c.Decode(msg.Payload())
				assert.NoError(t, err)
				decodedBlock, ok := decodedMsg.(*block.Block)
				assert.True(t, ok)
				h, err := decodedBlock.Hash()
				assert.NoError(t, err)
				assert.True(t, h.Equal(blockHash))
			case <-time.After(500 * time.Millisecond):
				t.Errorf("peer %d does not received msg", i)
			}
		}()
	}

	wgBlock.Wait()
}

func TestServer_ProcessSyncMessage(t *testing.T) {
	mainNodeAddr := ":30000"
	privKey, err := crypto.GeneratePrivateKey()
	assert.NoError(t, err)
	mainAddress := privKey.PublicKey().Address()
	assert.NoError(t, err)
	maxPeers := 7
	mainNode := tcp.NewTcpNode(privKey, mainNodeAddr, maxPeers)

	mainNode.Listen()
	time.Sleep(100 * time.Millisecond)

	numPeers := 7
	remoteNodes := make([]*tcp.TcpNode, numPeers)
	for i := 0; i < numPeers; i++ {
		time.Sleep(50 * time.Millisecond)
		go func() {
			remoteNode := tcp.GenerateTestTcpNode(t, fmt.Sprintf(":%d", 30001+i), maxPeers)
			remoteNode.Listen()

			time.Sleep(100 * time.Millisecond)
			remoteNodes[i] = remoteNode
			assert.NoError(t, remoteNode.Connect(mainNodeAddr))
		}()
	}

	t.Cleanup(func() {
		mainNode.Close()
		for _, n := range remoteNodes {
			n.Close()
		}
	})

	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, maxPeers, len(mainNode.Peers()))
	for i := 0; i < numPeers; i++ {
		assert.Equal(t, 1, len(remoteNodes[i].Peers()))
	}

	bc := core.GenerateTestBlockchain(t)
	mp := core.GenerateTestMempool(t, 100)
	sn := synchronizer.NewChainSynchronizer(mainAddress, mainNodeAddr, bc)
	sn.Start()
	defer sn.Stop()
	sc := synchronizer.NewDefaultSyncMessageCodec()
	gc := message.NewDefaultGossipMessageCodec()
	s := &Server{
		logger: util.LoggerWithPrefixes("Server"),
		state:  types.NewAtomicNumber[ServerState](Online),
	}
	s.Node = mainNode
	s.Chain = bc
	s.VirtualMemoryPool = mp
	s.Synchronizer = sn
	s.SyncMessageCodec = sc
	s.GossipMessageCodec = gc

	reqStatMsg := &synchronizer.RequestStatusMessage{}
	reqStatMsgPayload, err := sc.Encode(reqStatMsg)
	assert.NoError(t, err)
	fromAddr := util.RandomAddress()
	reqStatMsgRm := tcp.NewTcpRawMessage(fromAddr, reqStatMsgPayload)
	s.processSyncMessage(reqStatMsgRm)

	gb, err := bc.GetCurrentBlock()
	assert.NoError(t, err)
	gbh, err := gb.Hash()
	assert.NoError(t, err)

	select {
	case sm := <-sn.OutgoingMessage():
		dirMsg, ok := sm.(*synchronizer.DirectedMessage)
		assert.True(t, ok)
		assert.True(t, fromAddr.Equal(dirMsg.Address()))
		res, ok := dirMsg.Message.(*synchronizer.ResponseStatusMessage)
		assert.True(t, ok)
		assert.True(t, mainAddress.Equal(res.Address))
		assert.Equal(t, mainNodeAddr, res.NetAddr)
		assert.Equal(t, uint32(0), res.Version)
		assert.Equal(t, uint64(0), res.Height)
		assert.Equal(t, byte(synchronizer.Idle), res.State)
		assert.True(t, gbh.Equal(res.GenesisBlockHash))
		assert.True(t, gbh.Equal(res.CurrentBlockHash))
		s.processSynchronizerMessage(sm)
	case <-time.After(500 * time.Millisecond):
		t.Errorf("peer main does not received msg\n")
	}

	// for broadcast RequestStatusMessage
	s.Synchronizer.NotifyChainLagging()

	select {
	case sm := <-sn.OutgoingMessage():
		dirMsg, ok := sm.(*synchronizer.BroadcastMessage)
		assert.True(t, ok)
		assert.True(t, types.Address{}.Equal(dirMsg.Address()))
		_, ok = dirMsg.Message.(*synchronizer.RequestStatusMessage)
		assert.True(t, ok)
		s.processSynchronizerMessage(sm)
	case <-time.After(500 * time.Millisecond):
		t.Errorf("peer main does not received msg\n")
	}

	<-time.After(100 * time.Millisecond)

	wgReqStat := new(sync.WaitGroup)
	wgReqStat.Add(numPeers)

	for i, remoteNode := range remoteNodes {
		go func() {
			defer wgReqStat.Done()
			select {
			case msg := <-remoteNode.ConsumeRawMessage():
				payload := msg.Payload()
				assert.NotEmpty(t, payload)
				assert.Equal(t, byte(message.MessageSyncType), payload[0])
				assert.Equal(t, byte(synchronizer.MessageSyncSubTypeRequestStatus), payload[1])
				decodedMsg, err := sc.Decode(msg.Payload())
				assert.NoError(t, err)
				_, ok := decodedMsg.(*synchronizer.RequestStatusMessage)
				assert.True(t, ok)
			case <-time.After(500 * time.Millisecond):
				t.Errorf("peer %d does not received msg", i)
			}
		}()
	}

	wgReqStat.Wait()
}
