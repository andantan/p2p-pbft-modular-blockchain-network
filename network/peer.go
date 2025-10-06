package network

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/codec"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/network/message"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/types"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/util"
	"github.com/go-kit/log"
	"io"
	"net"
	"sync"
	"time"
)

type Peer interface {
	PublicKey() *crypto.PublicKey
	Address() types.Address
	NetAddr() string
	Handshake(*message.HandshakeMessage) (*message.HandshakeMessage, error)
	Send(payload []byte) error
	Read()
	Close()
}

type TCPPeer struct {
	publickey crypto.PublicKey
	address   types.Address
	netAddr   string

	logger log.Logger

	conn net.Conn

	msgCh chan message.RawMessage
	delCh chan Peer

	closeCh   chan struct{}
	closeOnce sync.Once
}

func NewTCPPeer(conn net.Conn, msgCh chan message.RawMessage, delCh chan Peer) Peer {
	return &TCPPeer{
		logger:  util.LoggerWithPrefixes("Peer", conn.RemoteAddr().String()),
		conn:    conn,
		msgCh:   msgCh,
		delCh:   delCh,
		closeCh: make(chan struct{}),
	}
}

func (p *TCPPeer) PublicKey() *crypto.PublicKey {
	return &p.publickey
}

func (p *TCPPeer) Address() types.Address {
	return p.address
}

func (p *TCPPeer) NetAddr() string {
	return p.netAddr
}

func (p *TCPPeer) Handshake(our *message.HandshakeMessage) (*message.HandshakeMessage, error) {
	msgCh := make(chan *message.HandshakeMessage, 1)
	defer close(msgCh)
	errCh := make(chan error, 2) // Read, Write error
	defer close(errCh)

	// read message
	go p.readHandshakeMessage(msgCh, errCh)

	// write message
	go p.writeHandshakeMessage(our, errCh)

	select {
	case err := <-errCh:
		_ = p.conn.Close()
		return nil, err
	case remoteMsg := <-msgCh:
		p.publickey = *remoteMsg.PublicKey
		p.address = remoteMsg.PublicKey.Address()
		p.netAddr = remoteMsg.NetAddr

		return remoteMsg, nil
	case <-time.After(5 * time.Second):
		_ = p.conn.Close()
		return nil, fmt.Errorf("message-timeout")
	}
}

func (p *TCPPeer) readHandshakeMessage(msgCh chan<- *message.HandshakeMessage, errCh chan<- error) {
	msgLenBuf := make([]byte, 4)

	if _, err := io.ReadFull(p.conn, msgLenBuf); IsUnrecoverableTCPError(err) {
		errCh <- err
		return
	}

	msgLen := binary.BigEndian.Uint32(msgLenBuf)

	// sanity check threshold = 4KB
	if msgLen > 1<<12 {
		errCh <- fmt.Errorf("message payload too large: %d bytes", msgLen)
		return
	}

	msgBuf := make([]byte, msgLen)
	if _, err := io.ReadFull(p.conn, msgBuf); IsUnrecoverableTCPError(err) {
		errCh <- err
		return
	}

	h := new(message.HandshakeMessage)

	if err := codec.DecodeProto(msgBuf, h); err != nil {
		errCh <- err
		return
	}

	if err := h.Verify(); err != nil {
		errCh <- err
		return
	}

	msgCh <- h
}

func (p *TCPPeer) writeHandshakeMessage(our *message.HandshakeMessage, errCh chan<- error) {
	payload, err := codec.EncodeProto(our)

	if err != nil {
		errCh <- err
		return
	}

	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(payload)))

	if _, err = p.conn.Write(append(lenBuf, payload...)); err != nil {
		errCh <- err
	}
}

func (p *TCPPeer) Send(payload []byte) error {
	lenBuf := make([]byte, 4)
	payloadSize := uint32(len(payload))
	binary.BigEndian.PutUint32(lenBuf, payloadSize)

	if _, err := p.conn.Write(append(lenBuf, payload...)); IsUnrecoverableTCPError(err) {
		p.Close()
		return err
	}

	return nil
}

func (p *TCPPeer) Read() {
	defer p.Close()

	rCh := make(chan []byte)
	defer close(rCh)
	errCh := make(chan error, 1)
	defer close(errCh)

	ctx, cancel := context.WithCancel(context.Background())

	wg := new(sync.WaitGroup)
	wg.Add(1)

	defer func(cancel context.CancelFunc) {
		cancel()
		wg.Done()
	}(cancel)

	go p.readMessages(ctx, wg, rCh, errCh)

	for {
		select {
		case <-p.closeCh:
			return

		case <-errCh:
			return

		case msgBuf := <-rCh:
			p.msgCh <- message.RawMessage{
				From:    p.Address(),
				Payload: bytes.NewBuffer(msgBuf),
			}
		}
	}
}

func (p *TCPPeer) readMessages(
	ctx context.Context,
	wg *sync.WaitGroup,
	rawCh chan<- []byte,
	errCh chan<- error,
) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		msgLenBuf := make([]byte, 4)
		if _, err := io.ReadFull(p.conn, msgLenBuf); IsUnrecoverableTCPError(err) {
			errCh <- err
			return
		}

		msgLen := binary.BigEndian.Uint32(msgLenBuf)
		if msgLen == 0 {
			// tcp heartbeat
			continue
		}

		msgBuf := make([]byte, msgLen)
		if _, err := io.ReadFull(p.conn, msgBuf); IsUnrecoverableTCPError(err) {
			errCh <- err
			return
		}

		rawCh <- msgBuf
	}
}

func (p *TCPPeer) Close() {
	p.closeOnce.Do(func() {
		_ = p.logger.Log("msg", "closing connection", "address", p.address.ShortString(8), "net-addr", p.NetAddr)

		close(p.closeCh)
		_ = p.conn.Close()
		if p.delCh != nil {
			p.delCh <- p
		}
	})
}
