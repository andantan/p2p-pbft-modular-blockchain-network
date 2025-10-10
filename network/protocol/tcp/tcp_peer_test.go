package tcp

import (
	"github.com/andantan/modular-blockchain/crypto"
	"github.com/stretchr/testify/assert"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

func TestTCPPeer_Handshake(t *testing.T) {
	connA, connB := net.Pipe()

	peerA := NewTcpPeer(connA)
	defer peerA.Close()
	peerB := NewTcpPeer(connB)
	defer peerB.Close()

	wg := new(sync.WaitGroup)
	wg.Add(2)

	var (
		MsgFromA *TcpHandshakeMessage
		MsgFromB *TcpHandshakeMessage
	)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		var err error

		privKeyA, err := crypto.GeneratePrivateKey()
		assert.NoError(t, err)
		msgA := NewTcpHandshakeMessage(privKeyA.PublicKey(), "addr_A")
		assert.NotNil(t, msgA.PublicKey)
		assert.NotNil(t, msgA.NetAddr)
		assert.NoError(t, msgA.Sign(privKeyA))
		assert.NotNil(t, msgA.Signature)

		MsgFromB, err = peerA.handshake(msgA)
		assert.NotNil(t, MsgFromB)
		assert.NoError(t, err)
	}(wg)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		var err error
		privKeyB, err := crypto.GeneratePrivateKey()
		assert.NoError(t, err)
		msgB := NewTcpHandshakeMessage(privKeyB.PublicKey(), "addr_B")
		assert.NotNil(t, msgB.PublicKey)
		assert.NotNil(t, msgB.NetAddr)
		assert.NoError(t, msgB.Sign(privKeyB))
		assert.NotNil(t, msgB.Signature)

		MsgFromA, err = peerB.handshake(msgB)
		assert.NotNil(t, MsgFromA)
		assert.NoError(t, err)
	}(wg)

	wg.Wait()

	assert.NoError(t, MsgFromA.Verify())
	assert.NoError(t, MsgFromB.Verify())

	expectedAddrA := MsgFromA.PublicKey.Address()
	assert.NotNil(t, peerA.Address())
	assert.Equal(t, expectedAddrA, peerB.Address())
	expectedAddrB := MsgFromB.PublicKey.Address()
	assert.NotNil(t, peerB.Address())
	assert.Equal(t, expectedAddrB, peerA.Address())

	assert.NotEqual(t, peerA.Address(), peerB.Address())
}

func TestTCPPeer_SendAndRead(t *testing.T) {
	connA, connB := net.Pipe()

	peerA := NewTcpPeer(connA)
	defer func() {
		peerA.Close()
		assert.True(t, peerA.state.Eq(Closed))
	}()
	peerB := NewTcpPeer(connB)
	defer func() {
		peerB.Close()
		assert.True(t, peerB.state.Eq(Closed))
	}()

	go peerB.Read()

	msgToSend := []byte("hello world")
	assert.NoError(t, peerA.Send(msgToSend))

	peerBMsgCh := peerB.ConsumeRawMessage()
	select {
	case rawMsg := <-peerBMsgCh:
		receivedMsg, err := io.ReadAll(rawMsg.Payload())
		assert.NoError(t, err)
		assert.Equal(t, msgToSend, receivedMsg)
	case <-time.After(1 * time.Second):
		t.Fatal("message was not received")
	}
}

func TestTCPPeer_Close(t *testing.T) {
	connA, connB := net.Pipe()

	peerA := NewTcpPeer(connA)
	defer func() {
		peerA.Close()
		assert.True(t, peerA.state.Eq(Closed))
	}()
	peerB := NewTcpPeer(connB)
	defer func() {
		peerB.Close()
		assert.True(t, peerB.state.Eq(Closed))
	}()

	assert.True(t, peerA.state.Eq(Initialized))
	assert.True(t, peerB.state.Eq(Initialized))

	wg := new(sync.WaitGroup)
	wg.Add(2)

	var (
		MsgFromA *TcpHandshakeMessage
		MsgFromB *TcpHandshakeMessage
	)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		var err error

		privKeyA, err := crypto.GeneratePrivateKey()
		assert.NoError(t, err)
		msgA := NewTcpHandshakeMessage(privKeyA.PublicKey(), "addr_A")
		assert.NotNil(t, msgA.PublicKey)
		assert.NotNil(t, msgA.NetAddr)
		assert.NoError(t, msgA.Sign(privKeyA))
		assert.NotNil(t, msgA.Signature)

		MsgFromB, err = peerA.handshake(msgA)
		assert.NotNil(t, MsgFromB)
		assert.NoError(t, err)
	}(wg)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		var err error
		privKeyB, err := crypto.GeneratePrivateKey()
		assert.NoError(t, err)
		msgB := NewTcpHandshakeMessage(privKeyB.PublicKey(), "addr_B")
		assert.NotNil(t, msgB.PublicKey)
		assert.NotNil(t, msgB.NetAddr)
		assert.NoError(t, msgB.Sign(privKeyB))
		assert.NotNil(t, msgB.Signature)

		MsgFromA, err = peerB.handshake(msgB)
		assert.NotNil(t, MsgFromA)
		assert.NoError(t, err)
	}(wg)

	wg.Wait()

	assert.True(t, peerA.state.Eq(Verified))
	assert.True(t, peerB.state.Eq(Verified))

	assert.NoError(t, MsgFromA.Verify())
	assert.NoError(t, MsgFromB.Verify())

	expectedAddrA := MsgFromA.PublicKey.Address()
	assert.NotNil(t, peerA.Address())
	assert.Equal(t, expectedAddrA, peerB.Address())
	expectedAddrB := MsgFromB.PublicKey.Address()
	assert.NotNil(t, peerB.Address())
	assert.Equal(t, expectedAddrB, peerA.Address())

	assert.NotEqual(t, peerA.Address(), peerB.Address())

	go peerA.Read()
	go peerB.Read()

	msgToSend := []byte("hello world")
	assert.NoError(t, peerA.Send(msgToSend))

	peerBMsgCh := peerB.ConsumeRawMessage()

	select {
	case rawMsg := <-peerBMsgCh:
		receivedMsg, err := io.ReadAll(rawMsg.Payload())
		assert.NoError(t, err)
		assert.Equal(t, msgToSend, receivedMsg)
	case <-time.After(1 * time.Second):
		t.Fatal("message was not received")
	}

	msgToSend2 := []byte("hello world2")
	assert.NoError(t, peerB.Send(msgToSend2))

	peerAMsgCh := peerA.ConsumeRawMessage()

	select {
	case rawMsg := <-peerAMsgCh:
		receivedMsg, err := io.ReadAll(rawMsg.Payload())
		assert.NoError(t, err)
		assert.Equal(t, msgToSend2, receivedMsg)
	case <-time.After(1 * time.Second):
		t.Fatal("message was not received")
	}
}
