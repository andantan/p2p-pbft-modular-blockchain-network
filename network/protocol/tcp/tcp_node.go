package tcp

import (
	"context"
	"errors"
	"github.com/andantan/modular-blockchain/codec"
	"github.com/andantan/modular-blockchain/crypto"
	"github.com/andantan/modular-blockchain/network/message"
	"github.com/andantan/modular-blockchain/types"
	"github.com/andantan/modular-blockchain/util"
	"github.com/go-kit/log"
	"net"
	"sync"
)

type TcpNode struct {
	logger log.Logger

	listenAddr string
	listener   net.Listener

	privKey *crypto.PrivateKey
	pubKey  *crypto.PublicKey
	address types.Address

	peerMap *types.AtomicMap[types.Address, *TcpPeer]

	outgoingMsgCh chan message.Raw
	closeCh       chan struct{}

	closeOnce sync.Once
}

func NewTcpNode(privKey *crypto.PrivateKey, listenAddr string) *TcpNode {
	t := &TcpNode{
		listenAddr:    listenAddr,
		privKey:       privKey,
		pubKey:        privKey.PublicKey(),
		address:       privKey.PublicKey().Address(),
		outgoingMsgCh: make(chan message.Raw, 1000),
		closeCh:       make(chan struct{}),
		peerMap:       types.NewAtomicMap[types.Address, *TcpPeer](),
	}

	t.logger = util.LoggerWithPrefixes("Node", "address", t.address.ShortString(8), "listen-addr", t.listenAddr)

	return t
}

func (n *TcpNode) Listen() {
	ln, err := net.Listen("tcp", n.listenAddr)

	if err != nil {
		panic(err)
	}

	_ = n.logger.Log("msg", "tcp listening", "net-addr", n.listenAddr)
	n.listener = ln

	go n.acceptLoop()
}

func (n *TcpNode) Connect(address string) error {
	_ = n.logger.Log("msg", "connecting to peer", "net_address", address)

	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}

	go n.handshakeAndValidate(conn)

	return nil
}

func (n *TcpNode) ConsumeRawMessage() <-chan message.Raw {
	return n.outgoingMsgCh
}

func (n *TcpNode) Broadcast(msg message.Message) error {
	payload, err := codec.EncodeProto(msg)
	if err != nil {
		return err
	}

	n.peerMap.Range(func(addr types.Address, peer *TcpPeer) bool {
		go func(p *TcpPeer) {
			if e := p.Send(payload); err != nil {
				_ = n.logger.Log("error", "failed to send message to peer", "peer_net_addr", p.NetAddr(), "peer_address", p.Address(), "err", e)
			}
		}(peer)
		return true
	})

	return nil
}

func (n *TcpNode) Disconnect(addr types.Address) {
	peer, ok := n.peerMap.Get(addr)
	if !ok {
		return
	}
	peer.Close()
}

func (n *TcpNode) Close() {
	n.closeOnce.Do(func() {
		_ = n.logger.Log("msg", "stopping tcp node")

		n.peerMap.Range(func(addr types.Address, peer *TcpPeer) bool {
			peer.Close()
			return true
		})

		close(n.closeCh)
		_ = n.listener.Close()
	})
}

func (n *TcpNode) acceptLoop() {
	_ = n.logger.Log("msg", "accept loop started")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	acceptCh := make(chan net.Conn)
	errCh := make(chan error, 1)

	go func(ctx context.Context) {
		defer close(acceptCh)
		defer close(errCh)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				conn, err := n.listener.Accept()
				if err != nil {
					errCh <- err
					return
				}
				acceptCh <- conn
			}
		}
	}(ctx)

	for {
		select {
		case conn, ok := <-acceptCh:
			if !ok {
				return
			}

			_ = n.logger.Log("msg", "new connection invoked. starting message", "net-addr", conn.RemoteAddr())
			go n.handshakeAndValidate(conn)

		case err := <-errCh:
			if errors.Is(err, net.ErrClosed) {
				_ = n.logger.Log("msg", "accept loop gracefully terminated")
			} else {
				_ = n.logger.Log("msg", "accept loop terminated due to error", "err", err)
			}
			return

		case <-n.closeCh:
			_ = n.logger.Log("msg", "server shutdown signal received, terminating accept loop")
			return
		}
	}
}

func (n *TcpNode) handshakeAndValidate(conn net.Conn) {
	select {
	case <-n.closeCh:
		return
	default:
	}

	_ = n.logger.Log("msg", "start handshake", "net-addr", conn.RemoteAddr())

	var (
		err      error
		ourId    = NewTcpHandshakeMessage(n.pubKey, n.listenAddr)
		remoteId *TcpHandshakeMessage
	)

	if err = ourId.Sign(n.privKey); err != nil {
		_ = conn.Close()
		_ = n.logger.Log("msg", "failed to sign our identity", "err", err)
		return
	}

	if err = ourId.Verify(); err != nil {
		_ = conn.Close()
		_ = n.logger.Log("msg", "failed to verify our identity", "err", err)
		return
	}

	peer := NewTcpPeer(conn)

	if remoteId, err = peer.handshake(ourId); err != nil {
		_ = n.logger.Log("msg", "failed to message peer", "err", err)
		return
	}

	remoteAddress := remoteId.PublicKey.Address()
	if !n.peerMap.PutIfNotExist(remoteAddress, peer) {
		// tie-breaking
		_ = n.logger.Log("msg", "tie-breaking: new commencing with existing peer", "net-addr", peer.NetAddr(), "address", peer.Address())
		isInbound := conn.LocalAddr() == n.listener.Addr()
		isLargerId := n.address.String() > remoteAddress.String()
		if isInbound && isLargerId {
			_ = n.logger.Log("msg", "tie-breaking: dropping inbound from lower ID peer")
			peer.Close()
		}

		return
	}

	_ = n.logger.Log("msg", "peer handshaking finished. send peer to newPeerCh", "net-addr", peer.NetAddr(), "address", peer.Address().ShortString(8))

	go peer.Read()
	go n.forwardMessages(peer)
}

func (n *TcpNode) forwardMessages(p *TcpPeer) {
	peerMsgCh := p.ConsumeRawMessage()

	for {
		select {
		case <-n.closeCh:
			return
		case msg, ok := <-peerMsgCh:
			if !ok {
				addr := p.Address()
				_ = n.logger.Log("msg", "closed peer channel", "peer_net_addr", p.NetAddr(), "peer_address", addr.ShortString(8))
				n.peerMap.Remove(addr)
				return
			}

			select {
			case n.outgoingMsgCh <- msg:
			case <-n.closeCh:
				return
			}
		}
	}
}
