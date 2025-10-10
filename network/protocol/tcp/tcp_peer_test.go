package tcp

import (
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/network"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/network/message"
	"github.com/stretchr/testify/assert"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

func TestTCPPeer_Handshake(t *testing.T) {
	connA, connB := net.Pipe()

	peerA := NewTcpPeer(connA, nil, nil)
	peerB := NewTcpPeer(connB, nil, nil)

	wg := new(sync.WaitGroup)
	wg.Add(2)

	var (
		MsgFromA network.Message
		MsgFromB network.Message
	)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		var err error

		privKeyA, err := crypto.GeneratePrivateKey()
		assert.NoError(t, err)
		msgA := NewTCPHandshakeMessage(privKeyA.PublicKey(), "addr_A")
		assert.NotNil(t, msgA.PublicKey)
		assert.NotNil(t, msgA.NetAddr)
		assert.NoError(t, msgA.Sign(privKeyA))
		assert.NotNil(t, msgA.Signature)

		MsgFromB, err = peerA.Handshake(msgA)
		assert.NotNil(t, MsgFromB)
		assert.NoError(t, err)
	}(wg)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		var err error
		privKeyB, err := crypto.GeneratePrivateKey()
		assert.NoError(t, err)
		msgB := NewTCPHandshakeMessage(privKeyB.PublicKey(), "addr_B")
		assert.NotNil(t, msgB.PublicKey)
		assert.NotNil(t, msgB.NetAddr)
		assert.NoError(t, msgB.Sign(privKeyB))
		assert.NotNil(t, msgB.Signature)

		MsgFromA, err = peerB.Handshake(msgB)
		assert.NotNil(t, MsgFromA)
		assert.NoError(t, err)
	}(wg)

	wg.Wait()

	aHand, ok := MsgFromA.(*TCPHandshakeMessage)
	assert.True(t, ok)
	bHand, ok := MsgFromB.(*TCPHandshakeMessage)
	assert.True(t, ok)

	assert.NoError(t, aHand.Verify())
	assert.NoError(t, bHand.Verify())

	expectedAddrA := aHand.PublicKey.Address()
	assert.NotNil(t, peerA.Address())
	assert.Equal(t, expectedAddrA, peerB.Address())
	expectedAddrB := bHand.PublicKey.Address()
	assert.NotNil(t, peerB.Address())
	assert.Equal(t, expectedAddrB, peerA.Address())

	assert.NotEqual(t, peerA.Address(), peerB.Address())
}

func TestTCPPeer_SendAndRead(t *testing.T) {
	msgCh := make(chan message.RawMessage, 1)
	delCh := make(chan *TcpPeer, 1)

	connA, connB := net.Pipe()

	peerA := NewTcpPeer(connA, msgCh, delCh)
	peerB := NewTcpPeer(connB, msgCh, delCh)

	go peerB.Read()

	msgToSend := []byte("hello world")
	assert.NoError(t, peerA.Send(msgToSend))

	select {
	case rawMsg := <-msgCh:
		receivedMsg, err := io.ReadAll(rawMsg.Payload)
		assert.NoError(t, err)
		assert.Equal(t, msgToSend, receivedMsg)
	case <-time.After(1 * time.Second):
		t.Fatal("message was not received")
	}
}

func TestTCPPeer_Close(t *testing.T) {
	msgCh := make(chan message.RawMessage)
	delCh := make(chan *TcpPeer, 1)

	connA, connB := net.Pipe()

	peerA := NewTcpPeer(connA, msgCh, delCh)
	peerB := NewTcpPeer(connB, msgCh, delCh)

	assert.True(t, peerA.state.Eq(Initialized))
	assert.True(t, peerB.state.Eq(Initialized))

	wg := new(sync.WaitGroup)
	wg.Add(2)

	var (
		MsgFromA network.Message
		MsgFromB network.Message
	)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		var err error

		privKeyA, err := crypto.GeneratePrivateKey()
		assert.NoError(t, err)
		msgA := NewTCPHandshakeMessage(privKeyA.PublicKey(), "addr_A")
		assert.NotNil(t, msgA.PublicKey)
		assert.NotNil(t, msgA.NetAddr)
		assert.NoError(t, msgA.Sign(privKeyA))
		assert.NotNil(t, msgA.Signature)

		MsgFromB, err = peerA.Handshake(msgA)
		assert.NotNil(t, MsgFromB)
		assert.NoError(t, err)
	}(wg)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		var err error
		privKeyB, err := crypto.GeneratePrivateKey()
		assert.NoError(t, err)
		msgB := NewTCPHandshakeMessage(privKeyB.PublicKey(), "addr_B")
		assert.NotNil(t, msgB.PublicKey)
		assert.NotNil(t, msgB.NetAddr)
		assert.NoError(t, msgB.Sign(privKeyB))
		assert.NotNil(t, msgB.Signature)

		MsgFromA, err = peerB.Handshake(msgB)
		assert.NotNil(t, MsgFromA)
		assert.NoError(t, err)
	}(wg)

	wg.Wait()

	assert.True(t, peerA.state.Eq(Verified))
	assert.True(t, peerB.state.Eq(Verified))

	aHand, ok := MsgFromA.(*TCPHandshakeMessage)
	assert.True(t, ok)
	bHand, ok := MsgFromB.(*TCPHandshakeMessage)
	assert.True(t, ok)

	assert.NoError(t, aHand.Verify())
	assert.NoError(t, bHand.Verify())

	expectedAddrA := aHand.PublicKey.Address()
	assert.NotNil(t, peerA.Address())
	assert.Equal(t, expectedAddrA, peerB.Address())
	expectedAddrB := bHand.PublicKey.Address()
	assert.NotNil(t, peerB.Address())
	assert.Equal(t, expectedAddrB, peerA.Address())

	assert.NotEqual(t, peerA.Address(), peerB.Address())

	go peerA.Read()
	go peerB.Read()

	msgToSend := []byte("hello world")
	assert.NoError(t, peerA.Send(msgToSend))

	select {
	case rawMsg := <-msgCh:
		receivedMsg, err := io.ReadAll(rawMsg.Payload)
		assert.NoError(t, err)
		assert.Equal(t, msgToSend, receivedMsg)
	case <-time.After(1 * time.Second):
		t.Fatal("message was not received")
	}

	go peerA.Close()
	go peerB.Close()

	receivedPeers := make(map[*TcpPeer]bool)
	for i := 0; i < 2; i++ {
		select {
		case terminatedPeer := <-delCh:
			receivedPeers[terminatedPeer] = true
		case <-time.After(1 * time.Second):
			t.Fatalf("all peer was not closed")
		}
	}

	assert.True(t, receivedPeers[peerA])
	assert.True(t, receivedPeers[peerB])
	assert.True(t, peerA.state.Eq(Closed))
	assert.True(t, peerB.state.Eq(Closed))
}
