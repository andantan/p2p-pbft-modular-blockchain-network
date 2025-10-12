package tcp

import (
	"fmt"
	"github.com/andantan/modular-blockchain/codec"
	"github.com/andantan/modular-blockchain/core/block"
	"github.com/stretchr/testify/assert"
	"net"
	"sync"
	"testing"
	"time"
)

func TestTCPNode_ConnectAndAccept(t *testing.T) {
	nodeANetAddr := ":4000"
	nodeA := GenerateTestTcpNode(t, nodeANetAddr, 1)

	nodeBNetAddr := ":5000"
	nodeB := GenerateTestTcpNode(t, nodeBNetAddr, 1)

	nodeA.Listen()
	go nodeA.acceptLoop()
	nodeB.Listen()
	go nodeB.acceptLoop()

	t.Cleanup(func() {
		nodeA.Close()
		nodeB.Close()
	})

	// delay for nodes starting
	time.Sleep(100 * time.Millisecond)

	assert.NoError(t, nodeB.Connect(nodeANetAddr))
	// delay for connect & handshake
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 1, nodeA.peerMap.Len())
	assert.Equal(t, 1, nodeB.peerMap.Len())

	testMsgContent := []byte("hello from B")
	peeraOfB, ok := nodeB.peerMap.Get(nodeA.address)
	assert.True(t, ok)
	assert.Equal(t, peeraOfB.Address(), nodeA.address)
	assert.NoError(t, peeraOfB.Send(testMsgContent))

	select {
	case rawMsg := <-nodeA.ConsumeRawMessage():
		assert.Equal(t, nodeB.address, rawMsg.From())
		assert.Equal(t, testMsgContent, rawMsg.Payload())
	case <-time.After(1 * time.Second):
		t.Fatal("Node A did not receive message from B")
	}

	nodeA.Disconnect(nodeB.address)
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 0, len(nodeA.Peers()))
	assert.Equal(t, 0, len(nodeB.Peers()))
}

func TestTCPNode_Stop(t *testing.T) {
	nodeNetAddr := ":7000"
	node := GenerateTestTcpNode(t, nodeNetAddr, 1)

	go node.Listen()
	// delay for node starting
	time.Sleep(100 * time.Millisecond)

	node.Close()
	// delay for node stopping
	time.Sleep(100 * time.Millisecond)

	ln, err := net.Listen("tcp", nodeNetAddr)

	// if listener closed normally, err must be nil
	assert.NoError(t, err)
	if err == nil {
		_ = ln.Close()
	}
}

func TestTCPNode_TieBreaking(t *testing.T) {
	nodeANetAddr := ":7000"
	nodeA := GenerateTestTcpNode(t, nodeANetAddr, 1)

	nodeBNetAddr := ":8000"
	nodeB := GenerateTestTcpNode(t, nodeBNetAddr, 1)

	go nodeA.Listen()
	go nodeB.Listen()
	t.Cleanup(func() {
		nodeA.Close()
		nodeB.Close()
	})
	// delay for nodeA, B starting
	time.Sleep(100 * time.Millisecond)

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func() {
		defer wg.Done()
		assert.NoError(t, nodeA.Connect(nodeBNetAddr))
	}()

	go func() {
		defer wg.Done()
		assert.NoError(t, nodeB.Connect(nodeANetAddr))
	}()

	wg.Wait()

	// delay for nodeA, B handshaking
	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, len(nodeA.Peers()), 1)
	assert.Equal(t, len(nodeB.Peers()), 1)
}

func TestTCPNode_MaxReached(t *testing.T) {
	maxPeers := 3
	mainNodeAddr := ":20000"
	mainNode := GenerateTestTcpNode(t, mainNodeAddr, maxPeers)

	mainNode.Listen()
	time.Sleep(100 * time.Millisecond)

	numPeers := 3
	remoteNodes := make([]*TcpNode, numPeers)
	for i := 0; i < numPeers; i++ {
		time.Sleep(50 * time.Millisecond)
		go func() {
			remoteNode := GenerateTestTcpNode(t, fmt.Sprintf(":%d", 20001+i), maxPeers)
			remoteNode.Listen()

			time.Sleep(100 * time.Millisecond)
			remoteNodes[i] = remoteNode
			assert.NoError(t, remoteNode.Connect(mainNodeAddr))
		}()
	}

	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, maxPeers, len(mainNode.Peers()))
	for i := 0; i < numPeers; i++ {
		assert.Equal(t, 1, len(remoteNodes[i].Peers()))
	}

	t.Cleanup(func() {
		mainNode.Close()
		for _, n := range remoteNodes {
			n.Close()
		}
	})

	lateNode := GenerateTestTcpNode(t, ":22222", maxPeers)

	t.Cleanup(func() {
		lateNode.Close()
	})

	lateNode.Listen()
	time.Sleep(100 * time.Millisecond)

	assert.NoError(t, lateNode.Connect(mainNodeAddr))
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 0, len(lateNode.Peers()))

	for i := 0; i < numPeers; i++ {
		assert.NoError(t, lateNode.Connect(remoteNodes[i].listenAddr))
	}
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 3, len(lateNode.Peers()))
}

func TestTCPNode_Broadcast(t *testing.T) {
	mainNodeAddr := ":10000"
	maxPeers := 7
	node := GenerateTestTcpNode(t, mainNodeAddr, maxPeers)

	node.Listen()
	time.Sleep(100 * time.Millisecond)

	numPeers := 7
	remoteNodes := make([]*TcpNode, numPeers)
	for i := 0; i < numPeers; i++ {
		time.Sleep(50 * time.Millisecond)
		go func() {
			remoteNode := GenerateTestTcpNode(t, fmt.Sprintf(":%d", 10001+i), maxPeers)
			remoteNode.Listen()

			time.Sleep(100 * time.Millisecond)
			remoteNodes[i] = remoteNode
			assert.NoError(t, remoteNode.Connect(mainNodeAddr))
		}()
	}

	t.Cleanup(func() {
		node.Close()
		for _, n := range remoteNodes {
			n.Close()
		}
	})

	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, maxPeers, len(node.Peers()))

	tx := block.GenerateRandomTestTransaction(t)
	payload, err := codec.EncodeProto(tx)
	assert.NoError(t, err)
	assert.NoError(t, node.Broadcast(payload))

	wg := new(sync.WaitGroup)
	wg.Add(numPeers)

	for _, remoteNode := range remoteNodes {
		go func(n *TcpNode) {
			defer wg.Done()
			select {
			case msg := <-n.ConsumeRawMessage():
				decodedTx := new(block.Transaction)
				assert.NoError(t, codec.DecodeProto(msg.Payload(), decodedTx))
				assert.NotEmpty(t, msg.Payload())
			case <-time.After(500 * time.Millisecond):
				t.Errorf("peer %s does not received msg", n.address.ShortString(8))
			}
		}(remoteNode)
	}

	wg.Wait()
}
