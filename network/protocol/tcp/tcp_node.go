package tcp

import (
	"context"
	"errors"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/network"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/network/message"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/types"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/util"
	"github.com/go-kit/log"
	"net"
)

type TcpNode struct {
	logger log.Logger

	listenAddr string
	listener   net.Listener

	privKey crypto.PrivateKey
	pubKey  crypto.PublicKey
	address types.Address

	messageCh chan message.RawMessage
	newPeerCh chan *TcpPeer
	delPeerCh chan *TcpPeer
	stopCh    chan struct{}

	peerSet *types.AtomicSet[types.Address]
}

func NewTcpNode(
	privKey *crypto.PrivateKey,
	listenAddr string,
	msgCh chan message.RawMessage,
	newPeerCh chan *TcpPeer,
	delPeerCh chan *TcpPeer,
	stopCh chan struct{},
) *TcpNode {
	t := &TcpNode{
		listenAddr: listenAddr,
		privKey:    *privKey,
		pubKey:     *privKey.PublicKey(),
		address:    privKey.PublicKey().Address(),
		messageCh:  msgCh,
		newPeerCh:  newPeerCh,
		delPeerCh:  delPeerCh,
		stopCh:     stopCh,
		peerSet:    types.NewAtomicSet[types.Address](),
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

		case <-n.stopCh:
			_ = n.logger.Log("msg", "server shutdown signal received, terminating accept loop")
			return
		}
	}
}

func (n *TcpNode) handshakeAndValidate(conn net.Conn) {
	_ = n.logger.Log("msg", "start handshake", "net-addr", conn.RemoteAddr())

	var (
		err      error
		ourId    = NewTCPHandshakeMessage(&n.pubKey, n.listenAddr)
		remoteId network.Message
	)

	if err = ourId.Sign(&n.privKey); err != nil {
		_ = conn.Close()
		_ = n.logger.Log("msg", "failed to sign our identity", "err", err)
		return
	}

	if err = ourId.Verify(); err != nil {
		_ = conn.Close()
		_ = n.logger.Log("msg", "failed to verify our identity", "err", err)
		return
	}

	peer := NewTcpPeer(conn, n.messageCh, n.delPeerCh)

	if remoteId, err = peer.Handshake(ourId); err != nil {
		_ = n.logger.Log("msg", "failed to message peer", "err", err)
		return
	}

	remoteHand, ok := remoteId.(*TCPHandshakeMessage)

	if !ok {
		_ = n.logger.Log("msg", "invalid remote handshake message type for TCPHandshakeMessage")
		_ = conn.Close()
		return
	}

	remoteAddress := remoteHand.PublicKey.Address()
	if !n.peerSet.PutIfNotExist(remoteAddress) {
		// tie-breaking
		_ = n.logger.Log("msg", "tie-breaking: new commencing with existing peer", "net-addr", peer.NetAddr(), "address", peer.Address())
		isInbound := conn.LocalAddr() == n.listener.Addr()
		isLargerId := n.address.String() > remoteAddress.String()
		if isInbound && isLargerId {
			_ = n.logger.Log("msg", "tie-breaking: dropping inbound from lower ID peer")
			_ = conn.Close()
			n.Remove(peer.Address())
		}

		return
	}

	_ = n.logger.Log("msg", "peer handshaking finished. send peer to newPeerCh", "net-addr", peer.NetAddr(), "address", peer.Address().ShortString(8))

	n.newPeerCh <- peer
}

func (n *TcpNode) Connect(address string) error {
	_ = n.logger.Log("msg", "connecting to peer", "address", address)

	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}

	go n.handshakeAndValidate(conn)

	return nil
}

func (n *TcpNode) Remove(addr types.Address) {
	n.peerSet.Remove(addr)
}

func (n *TcpNode) Stop() {
	_ = n.logger.Log("msg", "stopping tcp node")
	_ = n.listener.Close()
}
