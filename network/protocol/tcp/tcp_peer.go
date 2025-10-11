package tcp

import (
	"encoding/binary"
	"fmt"
	"github.com/andantan/modular-blockchain/codec"
	"github.com/andantan/modular-blockchain/crypto"
	"github.com/andantan/modular-blockchain/network/message"
	"github.com/andantan/modular-blockchain/types"
	"github.com/andantan/modular-blockchain/util"
	"github.com/go-kit/log"
	"io"
	"net"
	"sync"
	"time"
)

type TcpPeerState byte

const (
	Initialized TcpPeerState = iota
	Handshaking
	Verified
	Closed
)

type TcpPeer struct {
	logger log.Logger

	state *types.AtomicNumber[TcpPeerState]

	publickey *crypto.PublicKey
	address   types.Address
	netAddr   string
	conn      net.Conn

	msgCh chan message.Raw

	closeCh   chan struct{}
	closeOnce sync.Once
}

func NewTcpPeer(conn net.Conn) *TcpPeer {
	return &TcpPeer{
		logger:  util.LoggerWithPrefixes("Peer"),
		state:   types.NewAtomicNumber[TcpPeerState](Initialized),
		conn:    conn,
		msgCh:   make(chan message.Raw, 100),
		closeCh: make(chan struct{}),
	}
}

func (p *TcpPeer) Send(payload []byte) error {
	lenBuf := make([]byte, 4)
	payloadSize := uint32(len(payload))
	binary.BigEndian.PutUint32(lenBuf, payloadSize)

	if _, err := p.conn.Write(append(lenBuf, payload...)); IsUnrecoverableTCPError(err) {
		p.Close()
		return err
	}

	return nil
}

func (p *TcpPeer) Read() {
	defer p.Close()

	for {
		select {
		case <-p.closeCh:
			return
		default:
		}

		msgLenBuf := make([]byte, 4)
		if _, err := io.ReadFull(p.conn, msgLenBuf); IsUnrecoverableTCPError(err) {
			return
		}

		msgLen := binary.BigEndian.Uint32(msgLenBuf)
		if msgLen == 0 {
			continue // tcp heartbeat
		}

		msgBuf := make([]byte, msgLen)
		if _, err := io.ReadFull(p.conn, msgBuf); IsUnrecoverableTCPError(err) {
			return
		}

		msg := NewTcpRawMessage(p.address, msgBuf)

		select {
		case <-p.closeCh:
			return
		case p.msgCh <- msg:
		}
	}
}

func (p *TcpPeer) Close() {
	p.closeOnce.Do(func() {
		_ = p.logger.Log("msg", "closing connection", "address", p.address.ShortString(8), "net-addr", p.NetAddr)

		close(p.closeCh)
		_ = p.conn.Close()
		close(p.msgCh)
		p.state.Set(Closed)
	})
}

func (p *TcpPeer) PublicKey() *crypto.PublicKey {
	return p.publickey
}

func (p *TcpPeer) Address() types.Address {
	return p.address
}

func (p *TcpPeer) NetAddr() string {
	return p.netAddr
}

func (p *TcpPeer) ConsumeRawMessage() <-chan message.Raw {
	return p.msgCh
}

func (p *TcpPeer) handshake(our *TcpHandshakeMessage) (*TcpHandshakeMessage, error) {
	if !p.state.CompareAndSwap(Initialized, Handshaking) {
		return nil, fmt.Errorf("peer is not in initialized state or already handshaked")
	}

	select {
	case <-p.closeCh:
		return nil, fmt.Errorf("peer closed")
	default:
	}

	p.state.Set(Handshaking)

	msgCh := make(chan *TcpHandshakeMessage, 1)
	defer close(msgCh)
	errCh := make(chan error, 2) // Read, Write error
	defer close(errCh)

	_ = p.logger.Log("msg", "start handshake", "net-addr", p.conn.RemoteAddr())

	// read message
	go p.readHandshakeMessage(msgCh, errCh)

	// write message
	go p.sendHandshakeMessage(our, errCh)

	select {
	case <-p.closeCh:
		return nil, fmt.Errorf("peer closed")
	case err := <-errCh:
		p.Close()
		return nil, err
	case remoteMsg := <-msgCh:
		p.publickey = remoteMsg.PublicKey
		p.address = remoteMsg.PublicKey.Address()
		p.netAddr = remoteMsg.NetAddr
		_ = p.logger.Log("msg", "successed handshake", "net-addr", p.NetAddr(), "addr", p.address.ShortString(8))
		p.state.Set(Verified)
		return remoteMsg, nil
	case <-time.After(5 * time.Second):
		p.Close()
		err := fmt.Errorf("handshake-timeout")
		_ = p.logger.Log("msg", "failed handshake", "err", err)
		return nil, err
	}
}

func (p *TcpPeer) sendHandshakeMessage(our *TcpHandshakeMessage, errCh chan<- error) {
	select {
	case <-p.closeCh:
		return
	default:
	}

	payload, err := codec.EncodeProto(our)

	if err != nil {
		select {
		case <-p.closeCh:
		case errCh <- err:
		}
		return
	}

	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(payload)))

	if _, err = p.conn.Write(append(lenBuf, payload...)); err != nil {
		select {
		case <-p.closeCh:
		case errCh <- err:
		}
	} else {
		_ = p.logger.Log("msg", "send handshake message", "net-addr", p.conn.RemoteAddr())
	}
}

func (p *TcpPeer) readHandshakeMessage(msgCh chan<- *TcpHandshakeMessage, errCh chan<- error) {
	select {
	case <-p.closeCh:
		return
	default:
	}

	msgLenBuf := make([]byte, 4)

	if _, err := io.ReadFull(p.conn, msgLenBuf); IsUnrecoverableTCPError(err) {
		select {
		case <-p.closeCh:
		case errCh <- err:
		}
		return
	}

	msgLen := binary.BigEndian.Uint32(msgLenBuf)
	msgBuf := make([]byte, msgLen)
	if _, err := io.ReadFull(p.conn, msgBuf); IsUnrecoverableTCPError(err) {
		select {
		case <-p.closeCh:
		case errCh <- err:
		}
		return
	}

	h := new(TcpHandshakeMessage)
	if err := codec.DecodeProto(msgBuf, h); err != nil {
		select {
		case <-p.closeCh:
		case errCh <- err:
		}
		return
	}

	if err := h.Verify(); err != nil {
		select {
		case <-p.closeCh:
		case errCh <- err:
		}
		return
	}
	
	select {
	case <-p.closeCh:
	case msgCh <- h:
	}
}
