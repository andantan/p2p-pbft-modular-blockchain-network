package network

import (
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
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

	peerA := NewTCPPeer(connA, nil, nil)
	peerB := NewTCPPeer(connB, nil, nil)

	wg := new(sync.WaitGroup)
	wg.Add(2)

	var (
		remoteMsgFromA *message.HandshakeMessage
		remoteMsgFromB *message.HandshakeMessage
	)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		privKeyA, err := crypto.GeneratePrivateKey()
		assert.NoError(t, err)
		msgA := message.NewHandshakeMessage(privKeyA.PublicKey(), "addr_A")
		assert.NotNil(t, msgA.PublicKey)
		assert.NotNil(t, msgA.NetAddr)
		assert.NoError(t, msgA.Sign(privKeyA))
		assert.NotNil(t, msgA.Signature)

		remoteMsgFromB, err = peerA.Handshake(msgA)
		assert.NoError(t, err)
	}(wg)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		privKeyB, err := crypto.GeneratePrivateKey()
		assert.NoError(t, err)
		msgB := message.NewHandshakeMessage(privKeyB.PublicKey(), "addr_B")
		assert.NotNil(t, msgB.PublicKey)
		assert.NotNil(t, msgB.NetAddr)
		assert.NoError(t, msgB.Sign(privKeyB))
		assert.NotNil(t, msgB.Signature)

		remoteMsgFromA, err = peerB.Handshake(msgB)
		assert.NoError(t, err)
	}(wg)

	wg.Wait()

	assert.NoError(t, remoteMsgFromA.Verify())
	assert.NoError(t, remoteMsgFromB.Verify())

	expectedAddrB := remoteMsgFromB.PublicKey.Address()
	assert.Equal(t, expectedAddrB, peerA.Address())
	expectedAddrA := remoteMsgFromA.PublicKey.Address()
	assert.Equal(t, expectedAddrA, peerB.Address())

	assert.NotNil(t, peerA.Address())
	assert.NotNil(t, peerB.Address())
	assert.NotEqual(t, peerA.Address(), peerB.Address())
}

func TestTCPPeer_SendAndRead(t *testing.T) {
	msgCh := make(chan message.RawMessage, 1)
	delCh := make(chan Peer, 1)

	connA, connB := net.Pipe()

	peerA := NewTCPPeer(connA, msgCh, delCh)
	peerB := NewTCPPeer(connB, msgCh, delCh)

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
