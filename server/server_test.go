package server

import (
	"bytes"
	"fmt"
	"github.com/andantan/modular-blockchain/core"
	"github.com/andantan/modular-blockchain/core/block"
	"github.com/andantan/modular-blockchain/crypto"
	"github.com/andantan/modular-blockchain/network/consensus/pbft"
	"github.com/andantan/modular-blockchain/network/message"
	"github.com/andantan/modular-blockchain/network/protocol/tcp"
	"github.com/andantan/modular-blockchain/network/provider"
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
	privKey, err := crypto.GeneratePrivateKey()
	assert.NoError(t, err)
	mainNode := tcp.NewTcpNode(privKey, mainNodeAddr, maxPeers)

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

	ticker := time.NewTicker(50 * time.Millisecond)
	tried := 0
	for {
		tried++
		if maxPeers == len(mainNode.Peers()) {
			break
		}
		<-ticker.C
		if tried > 10 {
			t.Fatalf("too long time to connect all peers")
		}
	}
	ticker.Stop()

	for i := 0; i < numPeers; i++ {
		assert.Equal(t, 1, len(remoteNodes[i].Peers()))
	}

	bc := core.GenerateTestBlockchain(t)
	mp := core.GenerateTestMempool(t, 100)
	c := message.NewDefaultGossipMessageCodec()
	l := util.LoggerWithPrefixes("Server")
	s := &Server{
		state: types.NewAtomicNumber[ServerState](Online),
	}
	no := NewNetworkOptions().WithNode(mainNodeAddr, maxPeers, mainNode).WithGossipMessageCodec(c)
	bo := NewBlockchainOptions().WithChain(bc, time.Minute).WithVirtualMemoryPool(mp, 100)
	so := NewServerOptions(false).WithPrivateKey(privKey).WithLogger(l)
	s.NetworkOptions = no
	s.BlockchainOptions = bo
	s.ServerOptions = so

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
	b.Tail = &block.Tail{}
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
	l := util.LoggerWithPrefixes("Server")
	s := &Server{
		state: types.NewAtomicNumber[ServerState](Online),
	}
	no := NewNetworkOptions().WithNode(mainNodeAddr, maxPeers, mainNode).WithSyncMessageCodec(sc).WithSynchronizer(sn)
	no = no.WithGossipMessageCodec(gc)
	bo := NewBlockchainOptions().WithChain(bc, time.Minute).WithVirtualMemoryPool(mp, 100)
	so := NewServerOptions(false).WithPrivateKey(privKey).WithLogger(l)
	s.NetworkOptions = no
	s.ServerOptions = so
	s.BlockchainOptions = bo

	reqStatMsg := &synchronizer.RequestStatusMessage{}
	reqStatMsgPayload, err := sc.Encode(reqStatMsg)
	assert.NoError(t, err)
	fromAddr := util.RandomAddress()
	reqStatMsgRm := tcp.NewTcpRawMessage(fromAddr, reqStatMsgPayload)
	s.processRawMessage(reqStatMsgRm)

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

func TestServer_ProcessConsensusMessage(t *testing.T) {
	mainNodeAddr := ":20500"
	privKey, err := crypto.GeneratePrivateKey()
	assert.NoError(t, err)
	mainAddress := privKey.PublicKey().Address()
	assert.NotNil(t, mainAddress)
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
			remoteNode := tcp.GenerateTestTcpNode(t, fmt.Sprintf(":%d", 20501+i), maxPeers)
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

	mpCap := 100
	bc := core.GenerateTestBlockchain(t)
	mp := core.GenerateTestMempool(t, mpCap)
	pc := core.NewBlockProcessor(bc)
	gb := message.NewDefaultGossipMessageCodec()
	cd := pbft.NewPbftConsensusMessageCodec()
	cf := pbft.NewPbftConsensusEngineFactory()
	pp := GenerateMockPeerProvider()
	l := util.LoggerWithPrefixes("Server")

	s := &Server{
		state: types.NewAtomicNumber[ServerState](Online),
	}

	no := NewNetworkOptions().WithNode(mainNodeAddr, maxPeers, mainNode).WithPeerProvider(pp).WithGossipMessageCodec(gb)
	bo := NewBlockchainOptions().WithChain(bc, 0).WithVirtualMemoryPool(mp, mpCap).WithProcessor(pc)
	co := NewConsensusOptions().WithConsensusEngineFactory(cf).WithConsensusMessageCodec(cd)
	so := NewServerOptions(true).WithPrivateKey(privKey).WithLogger(l)

	s.NetworkOptions = no
	s.ServerOptions = so
	s.BlockchainOptions = bo
	s.ConsensusOptions = co

	assert.NoError(t, pp.Register(s.getOurInfo())) // 1 Validator

	prevHeader, err := bc.GetCurrentHeader()
	assert.NoError(t, err)
	genesisHash, err := prevHeader.Hash()
	assert.NoError(t, err)
	targetBlock := block.GenerateRandomTestBlockWithPrevHeaderAndKey(t, prevHeader, 1<<5, privKey)
	targetBlockHash, err := targetBlock.Hash()
	assert.NoError(t, err)
	prePrepareMsg := pbft.GenerateTestPbftPrePrepareMessageWithBlockAndKey(t, targetBlock, privKey, 0)
	prePrepareMsgPayload, err := cd.Encode(prePrepareMsg)
	assert.NoError(t, err)
	fromAddr := util.RandomAddress()
	prePrepareMsgRm := tcp.NewTcpRawMessage(fromAddr, prePrepareMsgPayload)
	s.processRawMessage(prePrepareMsgRm)

	select {
	case cm := <-s.ForwardConsensusEngineMessageCh:
		_, ok := cm.(*pbft.PbftPrePrepareMessage)
		assert.True(t, ok)
		s.processConsensusEngineMessage(cm)
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("forward message timeout")
	}

	assert.Equal(t, 1, len(s.ConsensusEngines))
	engine, ok := s.ConsensusEngines[targetBlock.Header.Height]
	t.Cleanup(func() {
		engine.Stop()
	})
	assert.True(t, ok)

	wgPrePrepare := new(sync.WaitGroup)
	wgPrePrepare.Add(numPeers)

	for i, remoteNode := range remoteNodes {
		go func() {
			defer wgPrePrepare.Done()
			select {
			case msg := <-remoteNode.ConsumeRawMessage():
				payload := msg.Payload()
				assert.NotEmpty(t, payload)
				assert.Equal(t, byte(message.MessageConsensusType), payload[0])
				assert.Equal(t, byte(pbft.MessagePbftConsensusSubTypePrePrepare), payload[1])
				decodedMsg, err := cd.Decode(msg.Payload())
				assert.NoError(t, err)
				decodedPrePrepare, ok := decodedMsg.(*pbft.PbftPrePrepareMessage)
				assert.True(t, ok)
				assert.Equal(t, uint64(0), decodedPrePrepare.View)
				assert.Equal(t, uint64(1), decodedPrePrepare.Sequence)
				b := decodedPrePrepare.Block
				hash, err := b.Hash()
				assert.NoError(t, err)
				assert.True(t, targetBlockHash.Equal(hash))
				assert.NoError(t, decodedPrePrepare.Verify())
			case <-time.After(500 * time.Millisecond):
				t.Errorf("peer %d does not received msg", i)
			}
		}()
	}

	wgPrePrepare.Wait()

	select {
	case cm := <-s.ForwardConsensusEngineMessageCh:
		decodedPrepare, ok := cm.(*pbft.PbftPrepareMessage)
		assert.True(t, ok)
		assert.Equal(t, uint64(0), decodedPrepare.View)
		assert.Equal(t, uint64(1), decodedPrepare.Sequence)
		assert.NoError(t, err)
		assert.True(t, targetBlockHash.Equal(decodedPrepare.BlockHash))
		assert.True(t, decodedPrepare.PublicKey.Equal(privKey.PublicKey()))
		assert.NoError(t, decodedPrepare.Verify())
		prepareMsgPayload, err := cd.Encode(decodedPrepare)
		assert.NoError(t, err)
		prepareMsgRm := tcp.NewTcpRawMessage(fromAddr, prepareMsgPayload)
		s.processRawMessage(prepareMsgRm)

	case <-time.After(200 * time.Millisecond):
		t.Fatalf("forward message timeout")
	}

	select {
	case cm := <-s.ForwardConsensusEngineMessageCh:
		decodedPrepare, ok := cm.(*pbft.PbftPrepareMessage)
		assert.True(t, ok)
		assert.Equal(t, uint64(0), decodedPrepare.View)
		assert.Equal(t, uint64(1), decodedPrepare.Sequence)
		assert.NoError(t, err)
		assert.True(t, targetBlockHash.Equal(decodedPrepare.BlockHash))
		assert.True(t, decodedPrepare.PublicKey.Equal(privKey.PublicKey()))
		assert.NoError(t, decodedPrepare.Verify())
		s.processConsensusEngineMessage(decodedPrepare)
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("forward message timeout")
	}

	wgPrepare := new(sync.WaitGroup)
	wgPrepare.Add(numPeers)

	for i, remoteNode := range remoteNodes {
		go func() {
			defer wgPrepare.Done()
			select {
			case msg := <-remoteNode.ConsumeRawMessage():
				payload := msg.Payload()
				assert.NotEmpty(t, payload)
				assert.Equal(t, byte(message.MessageConsensusType), payload[0])
				assert.Equal(t, byte(pbft.MessagePbftConsensusSubTypePrepare), payload[1])
				decodedMsg, err := cd.Decode(msg.Payload())
				assert.NoError(t, err)
				decodedPrepare, ok := decodedMsg.(*pbft.PbftPrepareMessage)
				assert.True(t, ok)
				assert.Equal(t, uint64(0), decodedPrepare.View)
				assert.Equal(t, uint64(1), decodedPrepare.Sequence)
				assert.True(t, targetBlockHash.Equal(decodedPrepare.BlockHash))
				assert.NoError(t, decodedPrepare.Verify())
			case <-time.After(500 * time.Millisecond):
				t.Errorf("peer %d does not received msg", i)
			}
		}()
	}

	wgPrepare.Wait()

	select {
	case cm := <-s.ForwardConsensusEngineMessageCh:
		decodedCommit, ok := cm.(*pbft.PbftCommitMessage)
		assert.True(t, ok)
		assert.Equal(t, uint64(0), decodedCommit.View)
		assert.Equal(t, uint64(1), decodedCommit.Sequence)
		assert.NoError(t, err)
		assert.True(t, targetBlockHash.Equal(decodedCommit.BlockHash))
		assert.True(t, decodedCommit.PublicKey.Equal(privKey.PublicKey()))
		assert.NoError(t, decodedCommit.Verify())
		commitDecodedPayload, err := cd.Encode(decodedCommit)
		assert.NoError(t, err)
		commitMsgRm := tcp.NewTcpRawMessage(fromAddr, commitDecodedPayload)
		s.processRawMessage(commitMsgRm)

	case <-time.After(200 * time.Millisecond):
		t.Fatalf("forward message timeout")
	}

	select {
	case cm := <-s.ForwardConsensusEngineMessageCh:
		decodedCommit, ok := cm.(*pbft.PbftCommitMessage)
		assert.True(t, ok)
		assert.Equal(t, uint64(0), decodedCommit.View)
		assert.Equal(t, uint64(1), decodedCommit.Sequence)
		assert.NoError(t, err)
		assert.True(t, targetBlockHash.Equal(decodedCommit.BlockHash))
		assert.True(t, decodedCommit.PublicKey.Equal(privKey.PublicKey()))
		assert.NoError(t, decodedCommit.Verify())
		s.processConsensusEngineMessage(decodedCommit)
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("forward message timeout")
	}

	wgCommit := new(sync.WaitGroup)
	wgCommit.Add(numPeers)

	for i, remoteNode := range remoteNodes {
		go func() {
			defer wgCommit.Done()
			select {
			case msg := <-remoteNode.ConsumeRawMessage():
				payload := msg.Payload()
				assert.NotEmpty(t, payload)
				assert.Equal(t, byte(message.MessageConsensusType), payload[0])
				assert.Equal(t, byte(pbft.MessagePbftConsensusSubTypeCommit), payload[1])
				decodedMsg, err := cd.Decode(msg.Payload())
				assert.NoError(t, err)
				decodedCommit, ok := decodedMsg.(*pbft.PbftCommitMessage)
				assert.True(t, ok)
				assert.Equal(t, uint64(0), decodedCommit.View)
				assert.Equal(t, uint64(1), decodedCommit.Sequence)
				assert.True(t, targetBlockHash.Equal(decodedCommit.BlockHash))
				assert.NoError(t, decodedCommit.Verify())
			case <-time.After(200 * time.Millisecond):
				t.Errorf("peer %d does not received msg", i)
			}
		}()
	}

	wgCommit.Wait()

	var fb *block.Block
	select {
	case fb = <-s.ForwardFinalizedBlockCh:
		assert.NoError(t, fb.Verify())
		assert.True(t, fb.IsConsented())
		s.processFinalizedBlock(fb)
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("forward message timeout")
	}

	wgFb := new(sync.WaitGroup)
	wgFb.Add(numPeers)

	for i, remoteNode := range remoteNodes {
		go func() {
			defer wgFb.Done()
			select {
			case msg := <-remoteNode.ConsumeRawMessage():
				payload := msg.Payload()
				assert.NotEmpty(t, payload)
				assert.Equal(t, byte(message.MessageGossipType), payload[0])
				assert.Equal(t, byte(message.MessageGossipSubTypeBlock), payload[1])
				decodedMsg, err := gb.Decode(msg.Payload())
				assert.NoError(t, err)
				decodedBlock, ok := decodedMsg.(*block.Block)
				assert.True(t, ok)
				assert.Equal(t, uint64(1), decodedBlock.Header.Height)
				decodedBlockHash, err := decodedBlock.Hash()
				assert.NoError(t, err)
				assert.True(t, targetBlockHash.Equal(decodedBlockHash))
				assert.True(t, genesisHash.Equal(decodedBlock.Header.PrevBlockHash))
				assert.NoError(t, decodedBlock.Verify())
				assert.True(t, decodedBlock.IsConsented())
				assert.Equal(t, 1, len(decodedBlock.Tail.CommitVotes))
			case <-time.After(200 * time.Millisecond):
				t.Errorf("peer %d does not received msg", i)
			}
		}()
	}

	wgFb.Wait()
}

func TestServer_ProposeBlockIfProposer_When_Proposer(t *testing.T) {
	mainNodeAddr := ":21500"
	privKey, err := crypto.GeneratePrivateKey()
	assert.NoError(t, err)
	mainAddress := privKey.PublicKey().Address()
	assert.NotNil(t, mainAddress)
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
			remoteNode := tcp.GenerateTestTcpNode(t, fmt.Sprintf(":%d", 21501+i), maxPeers)
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

	mpCap := 100
	bc := core.GenerateTestBlockchain(t)
	mp := core.GenerateTestMempool(t, mpCap)
	gb := message.NewDefaultGossipMessageCodec()
	cd := pbft.NewPbftConsensusMessageCodec()
	cl := pbft.NewPbftLeaderSelector()
	po := pbft.NewPbftProposer(privKey, bc, mp)
	pp := GenerateMockPeerProvider()
	l := util.LoggerWithPrefixes("Server")

	tx := block.GenerateRandomTestTransaction(t)
	assert.NoError(t, mp.Put(tx))

	s := &Server{
		state: types.NewAtomicNumber[ServerState](Online),
	}

	no := NewNetworkOptions().WithNode(mainNodeAddr, maxPeers, mainNode).WithPeerProvider(pp).WithGossipMessageCodec(gb)
	bo := NewBlockchainOptions().WithChain(bc, 0).WithVirtualMemoryPool(mp, mpCap)
	co := NewConsensusOptions().WithConsensusMessageCodec(cd).WithLeaderSelector(cl).WithProposer(po)
	so := NewServerOptions(true).WithPrivateKey(privKey).WithLogger(l)

	s.NetworkOptions = no
	s.ServerOptions = so
	s.BlockchainOptions = bo
	s.ConsensusOptions = co

	assert.NoError(t, pp.Register(s.getOurInfo())) // 1 Validator

	s.proposeBlockIfProposer()

	wgPrePrepare := new(sync.WaitGroup)
	wgPrePrepare.Add(numPeers)

	for i, remoteNode := range remoteNodes {
		go func() {
			defer wgPrePrepare.Done()
			select {
			case msg := <-remoteNode.ConsumeRawMessage():
				payload := msg.Payload()
				assert.NotEmpty(t, payload)
				assert.Equal(t, byte(message.MessageConsensusType), payload[0])
				assert.Equal(t, byte(pbft.MessagePbftConsensusSubTypePrePrepare), payload[1])
				decodedMsg, err := cd.Decode(msg.Payload())
				assert.NoError(t, err)
				decodedPrePrepare, ok := decodedMsg.(*pbft.PbftPrePrepareMessage)
				assert.True(t, ok)
				assert.Equal(t, uint64(0), decodedPrePrepare.View)
				assert.Equal(t, uint64(1), decodedPrePrepare.Sequence)
				assert.NoError(t, decodedPrePrepare.Verify())
			case <-time.After(500 * time.Millisecond):
				t.Errorf("peer %d does not received msg", i)
			}
		}()
	}

	wgPrePrepare.Wait()
}

func TestServer_ProposeBlockIfProposer_When_NotProposer(t *testing.T) {
	mainNodeAddr := ":22500"
	privKey, err := crypto.GeneratePrivateKey()
	assert.NoError(t, err)
	mainAddress := privKey.PublicKey().Address()
	assert.NotNil(t, mainAddress)
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
			remoteNode := tcp.GenerateTestTcpNode(t, fmt.Sprintf(":%d", 22501+i), maxPeers)
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

	mpCap := 100
	bc := core.GenerateTestBlockchain(t)
	mp := core.GenerateTestMempool(t, mpCap)
	gb := message.NewDefaultGossipMessageCodec()
	cd := pbft.NewPbftConsensusMessageCodec()
	cl := pbft.NewPbftLeaderSelector()
	po := pbft.NewPbftProposer(privKey, bc, mp)
	pp := GenerateMockPeerProvider()
	l := util.LoggerWithPrefixes("Server")

	tx := block.GenerateRandomTestTransaction(t)
	assert.NoError(t, mp.Put(tx))

	s := &Server{
		state: types.NewAtomicNumber[ServerState](Online),
	}

	no := NewNetworkOptions().WithNode(mainNodeAddr, maxPeers, mainNode).WithPeerProvider(pp).WithGossipMessageCodec(gb)
	bo := NewBlockchainOptions().WithChain(bc, 0).WithVirtualMemoryPool(mp, mpCap)
	co := NewConsensusOptions().WithConsensusMessageCodec(cd).WithLeaderSelector(cl).WithProposer(po)
	so := NewServerOptions(true).WithPrivateKey(privKey).WithLogger(l)

	s.NetworkOptions = no
	s.ServerOptions = so
	s.BlockchainOptions = bo
	s.ConsensusOptions = co

	assert.NoError(t, pp.Register(s.getOurInfo())) // 1 Validator

	temperPeerInfo := &provider.PeerInfo{
		Address:        util.RandomAddress().String(),
		NetAddr:        "tempering",
		Connections:    uint8(5),
		MaxConnections: uint8(16),
		Height:         uint64(0),
		IsValidator:    true,
	}
	pp.Peers[s.Address] = temperPeerInfo

	s.proposeBlockIfProposer()

	wgPrePrepare := new(sync.WaitGroup)
	wgPrePrepare.Add(numPeers)

	for _, remoteNode := range remoteNodes {
		go func() {
			defer wgPrePrepare.Done()
			select {
			case <-remoteNode.ConsumeRawMessage():
				t.Errorf("mainNode should not propose block")
			case <-time.After(500 * time.Millisecond):
				return
			}
		}()
	}

	wgPrePrepare.Wait()
}

func TestServer_ProcessCreatedTransaction(t *testing.T) {
	mainNodeAddr := ":23500"
	privKey, err := crypto.GeneratePrivateKey()
	assert.NoError(t, err)
	mainAddress := privKey.PublicKey().Address()
	assert.NotNil(t, mainAddress)
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
			remoteNode := tcp.GenerateTestTcpNode(t, fmt.Sprintf(":%d", 23501+i), maxPeers)
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

	mpCap := 100
	mp := core.GenerateTestMempool(t, mpCap)
	gb := message.NewDefaultGossipMessageCodec()
	as := GenerateMockApiServer()
	as.Start()
	t.Cleanup(func() {
		as.Stop()
	})
	l := util.LoggerWithPrefixes("Server")

	s := &Server{
		state: types.NewAtomicNumber[ServerState](Online),
	}

	no := NewNetworkOptions().WithNode(mainNodeAddr, maxPeers, mainNode).WithApiServer("api", as).WithGossipMessageCodec(gb)
	bo := NewBlockchainOptions().WithVirtualMemoryPool(mp, mpCap)
	so := NewServerOptions(true).WithPrivateKey(privKey).WithLogger(l)

	s.NetworkOptions = no
	s.ServerOptions = so
	s.BlockchainOptions = bo

	newTx := block.GenerateRandomTestTransaction(t)
	newTxHash, err := newTx.Hash()
	assert.NoError(t, err)
	s.processCreatedTransaction(newTx)

	wgTx := new(sync.WaitGroup)
	wgTx.Add(numPeers)

	for i, remoteNode := range remoteNodes {
		go func() {
			defer wgTx.Done()
			select {
			case msg := <-remoteNode.ConsumeRawMessage():
				payload := msg.Payload()
				assert.NotEmpty(t, payload)
				assert.Equal(t, byte(message.MessageGossipType), payload[0])
				assert.Equal(t, byte(message.MessageGossipSubTypeTransaction), payload[1])
				decodedMsg, err := gb.Decode(msg.Payload())
				assert.NoError(t, err)
				decodedTransaction, ok := decodedMsg.(*block.Transaction)
				assert.True(t, ok)
				decodedHash, err := decodedTransaction.Hash()
				assert.NoError(t, err)
				assert.True(t, newTxHash.Equal(decodedHash))
				assert.NoError(t, decodedTransaction.Verify())
				assert.True(t, bytes.Equal(newTx.Data, decodedTransaction.Data))
				assert.True(t, newTx.From.Equal(decodedTransaction.From))
				assert.Equal(t, newTx.Nonce, decodedTransaction.Nonce)
			case <-time.After(500 * time.Millisecond):
				t.Errorf("peer %d does not received msg", i)
			}
		}()
	}

	wgTx.Wait()
}
