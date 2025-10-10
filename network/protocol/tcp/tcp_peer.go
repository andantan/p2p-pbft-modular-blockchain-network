package tcp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/codec"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/network"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/network/message"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/types"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/util"
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

	publickey crypto.PublicKey
	address   types.Address
	netAddr   string

	state *types.AtomicNumber[TcpPeerState]
	conn  net.Conn

	msgCh chan message.RawMessage
	delCh chan *TcpPeer

	closeCh   chan struct{}
	closeOnce sync.Once
}

func NewTcpPeer(conn net.Conn, msgCh chan message.RawMessage, delCh chan *TcpPeer) *TcpPeer {
	return &TcpPeer{
		logger:  util.LoggerWithPrefixes("Peer"),
		state:   types.NewAtomicNumber[TcpPeerState](Initialized),
		conn:    conn,
		msgCh:   msgCh,
		delCh:   delCh,
		closeCh: make(chan struct{}),
	}
}

func (p *TcpPeer) PublicKey() *crypto.PublicKey {
	return &p.publickey
}

func (p *TcpPeer) Address() types.Address {
	return p.address
}

func (p *TcpPeer) NetAddr() string {
	return p.netAddr
}

func (p *TcpPeer) Handshake(our network.Message) (network.Message, error) {
	if !p.state.CompareAndSwap(Initialized, Handshaking) {
		return nil, fmt.Errorf("peer is not in initialized state or already handshaked")
	}

	select {
	case <-p.closeCh:
		return nil, fmt.Errorf("peer closed")
	default:
	}

	p.state.Set(Handshaking)
	ourhands, ok := our.(*TCPHandshakeMessage)

	if !ok {
		return nil, errors.New("invalid our handshake message type for TCPHandshakeMessage")
	}

	msgCh := make(chan *TCPHandshakeMessage, 1)
	defer close(msgCh)
	errCh := make(chan error, 2) // Read, Write error
	defer close(errCh)

	_ = p.logger.Log("msg", "start handshake", "net-addr", p.conn.RemoteAddr())

	// read message
	go p.readHandshakeMessage(msgCh, errCh)

	// write message
	go p.writeHandshakeMessage(ourhands, errCh)

	select {
	case <-p.closeCh:
		return nil, fmt.Errorf("peer closed")
	case err := <-errCh:
		p.Close()
		return nil, err
	case remoteMsg := <-msgCh:
		p.publickey = *remoteMsg.PublicKey
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

func (p *TcpPeer) readHandshakeMessage(msgCh chan<- *TCPHandshakeMessage, errCh chan<- error) {
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

	h := new(TCPHandshakeMessage)

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

	_ = p.logger.Log("msg", "received and verified handshake messaage", "net-addr", h.NetAddr, "addr", h.PublicKey.Address().ShortString(8))

	select {
	case <-p.closeCh:
	case msgCh <- h:
	}
}

func (p *TcpPeer) writeHandshakeMessage(our *TCPHandshakeMessage, errCh chan<- error) {
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

		msg := message.RawMessage{
			From:    p.Address(),
			Payload: bytes.NewBuffer(msgBuf),
		}

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
		if p.state.Eq(Verified) && p.delCh != nil {
			p.delCh <- p
		}
		p.state.Set(Closed)
	})
}
