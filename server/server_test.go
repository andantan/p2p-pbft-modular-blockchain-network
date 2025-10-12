package server

import (
	"fmt"
	"github.com/andantan/modular-blockchain/core"
	"github.com/andantan/modular-blockchain/core/block"
	"github.com/andantan/modular-blockchain/network/message"
	"github.com/andantan/modular-blockchain/network/protocol/tcp"
	"github.com/andantan/modular-blockchain/types"
	"github.com/andantan/modular-blockchain/util"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestServer_ProcessGossipMessage(t *testing.T) {
	mainNodeAddr := ":10500"
	maxPeers := 7
	mainNode := tcp.GenerateTestTcpNode(t, mainNodeAddr, maxPeers)

	mainNode.Listen()
	time.Sleep(100 * time.Millisecond)

	numPeers := 7
	remoteNodes := make([]*tcp.TcpNode, numPeers)
	for i := 0; i < numPeers; i++ {
		time.Sleep(50 * time.Millisecond)
		go func() {
			remoteNode := tcp.GenerateTestTcpNode(t, fmt.Sprintf(":%d", 10001+i), maxPeers)
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
	s.processMessage(txRm)

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
	s.processMessage(blockRm)

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
