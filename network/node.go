package network

import (
	"context"
	"errors"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/network/message"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/types"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/util"
	"github.com/go-kit/log"
	"net"
)

type Node interface {
	Listen()
	Connect(string) error
	Remove(Peer)
	Stop()
}

type TCPNode struct {
	logger log.Logger

	listenAddr string
	listener   net.Listener

	privKey crypto.PrivateKey
	pubKey  crypto.PublicKey
	address types.Address

	messageCh chan message.RawMessage
	newPeerCh chan Peer
	delPeerCh chan Peer
	stopCh    chan struct{}

	peerSet *types.AtomicSet[types.Address]
}

func NewTCPNode(
	privKey *crypto.PrivateKey,
	listenAddr string,
	msgCh chan message.RawMessage,
	newPeerCh chan Peer,
	delPeerCh chan Peer,
	stopCh chan struct{},
) *TCPNode {
	t := &TCPNode{
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

	t.logger = util.LoggerWithPrefixes("node", "address", t.address.ShortString(8), "listen-addr", t.listenAddr)

	return t
}

func (n *TCPNode) Listen() {
	ln, err := net.Listen("tcp", n.listenAddr)

	if err != nil {
		panic(err)
	}

	_ = n.logger.Log("msg", "tcp listening", "net-addr", n.listenAddr)
	n.listener = ln

	go n.acceptLoop()
}

func (n *TCPNode) acceptLoop() {
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

			_ = n.logger.Log("msg", "new connection invoked. starting handshake", "net-addr", conn.RemoteAddr())
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

func (n *TCPNode) handshakeAndValidate(conn net.Conn) {
	var (
		err      error
		ourId    = message.NewHandshakeMessage(&n.pubKey, n.listenAddr)
		remoteId *message.HandshakeMessage
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

	peer := NewTCPPeer(conn, n.messageCh, n.delPeerCh)

	if remoteId, err = peer.Handshake(ourId); err != nil {
		_ = n.logger.Log("msg", "failed to handshake peer", "err", err)
		return
	}

	remoteAddress := remoteId.PublicKey.Address()
	if !n.peerSet.PutIfNotExist(remoteAddress) {
		// tie-breaking
		_ = n.logger.Log("msg", "tie-breaking: new commencing with existing peer", "net-addr", peer.NetAddr(), "address", peer.Address())
		isInbound := conn.LocalAddr() == n.listener.Addr()
		isLargerId := n.address.String() > remoteAddress.String()
		if isInbound && isLargerId {
			_ = n.logger.Log("msg", "tie-breaking: dropping inbound from lower ID peer")
			_ = conn.Close()
		}

		return
	}

	_ = n.logger.Log("msg", "peer handshaking finished. send peer to newPeerCh", "net-addr", peer.NetAddr(), "address", peer.Address().ShortString(8))

	n.newPeerCh <- peer
}

func (n *TCPNode) Connect(address string) error {
	_ = n.logger.Log("msg", "connecting to peer", "address", address)

	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}

	go n.handshakeAndValidate(conn)

	return nil
}

func (n *TCPNode) Remove(peer Peer) {
	n.peerSet.Remove(peer.Address())
}

func (n *TCPNode) Stop() {
	_ = n.logger.Log("msg", "stopping tcp node")
	_ = n.listener.Close()
}
