package server

import (
	"errors"
	"fmt"
	"github.com/andantan/modular-blockchain/core"
	"github.com/andantan/modular-blockchain/core/block"
	"github.com/andantan/modular-blockchain/crypto"
	"github.com/andantan/modular-blockchain/network/message"
	"github.com/andantan/modular-blockchain/network/provider"
	"github.com/andantan/modular-blockchain/network/synchronizer"
	"github.com/andantan/modular-blockchain/types"
	"github.com/andantan/modular-blockchain/util"
	"github.com/go-kit/log"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"
)

type ServerState byte

const (
	Initialized ServerState = iota
	Bootstrap
	Online
	Offline
)

func (s ServerState) String() string {
	switch s {
	case Initialized:
		return "Initialized"
	case Bootstrap:
		return "BOOTSTRAP"
	case Online:
		return "ONLINE"
	case Offline:
		return "OFFLINE"
	default:
		return "UNKNOWN"
	}
}

type Server struct {
	ServerOptions
	NetworkOptions
	BlockchainOptions
	ConsensusOptions

	logger log.Logger

	PublicKey *crypto.PublicKey
	Address   types.Address

	state *types.AtomicNumber[ServerState]

	closeCh   chan struct{}
	closeOnce sync.Once
	wg        sync.WaitGroup
}

func NewServer(so ServerOptions, no NetworkOptions, bo BlockchainOptions, co ConsensusOptions) (*Server, error) {
	if !so.IsFulFilled() {
		panic("ServerOptions is not fulfilled")
	}

	if !no.IsFulFilled() {
		panic("ServerOptions is not fulfilled")
	}

	if !bo.IsFulFilled() {
		panic("BlockchainOptions is not fulfilled")
	}

	if !co.IsFulFilled() {
		panic("ConsensusOptions is not fulfilled")
	}

	return &Server{
		ServerOptions:     so,
		NetworkOptions:    no,
		BlockchainOptions: bo,
		ConsensusOptions:  co,
		logger:            util.LoggerWithPrefixes("Server"),
		PublicKey:         so.PrivateKey.PublicKey(),
		Address:           so.PrivateKey.PublicKey().Address(),
		state:             types.NewAtomicNumber[ServerState](Initialized),
		closeCh:           make(chan struct{}),
	}, nil
}

func (s *Server) Start(wg *sync.WaitGroup) {
	defer wg.Done()
	s.state.Set(Bootstrap)

	s.Node.Listen()
	s.Chain.Bootstrap()
	s.ApiServer.Start()

	go s.manageConnections()
	go s.heartbeatLoop()
	go s.logLoop()

	if err := s.Provider.Register(s.getOurInfo()); err != nil {
		panic(err)
	}

	s.loop()
}

func (s *Server) Stop() {
	s.closeOnce.Do(func() {
		s.Node.Close()
		_ = s.logger.Log("msg", "node is shutting down")
		s.ApiServer.Stop()
		_ = s.logger.Log("msg", "api server is shutting down")

		for _, engine := range s.ConsensusEngines {
			engine.Stop()
		}
		_ = s.logger.Log("msg", "consensus engines are shutting down")

		close(s.closeCh)
	})
}

func (s *Server) getOurInfo() *provider.PeerInfo {
	return &provider.PeerInfo{
		Address:        s.Address.String(),
		NetAddr:        s.ListenAddr,
		Connections:    uint8(len(s.Node.Peers())),
		MaxConnections: uint8(s.MaxPeers),
		Height:         s.Chain.GetCurrentHeight(),
		IsValidator:    s.IsValidator,
	}
}

func (s *Server) manageConnections() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.closeCh:
			return
		case <-ticker.C:
			if len(s.Node.Peers()) < s.MaxPeers {
				s.discoverPeer()
			} else {
				s.churningPeer()
			}
		}
	}
}

func (s *Server) discoverPeer() {
	currentPeers := s.Node.Peers()
	if s.MaxPeers <= len(currentPeers) {
		return
	}

	peerCandidates, err := s.Provider.DiscoverPeers()
	if err != nil {
		_ = s.logger.Log("msg", "failed to discover peers", "err", err)
	}

	if len(peerCandidates) == 0 {
		return
	}

	connectedPeers := make(map[types.Address]struct{})
	for _, peer := range currentPeers {
		connectedPeers[peer] = struct{}{}
	}

	filter := make([]*provider.PeerInfo, 0)
	for _, candidate := range peerCandidates {
		addr, err := types.AddressFromHexString(candidate.Address)
		if err != nil {
			continue
		}

		if _, ok := connectedPeers[addr]; ok {
			continue
		}

		if s.Address.Equal(addr) {
			continue
		}

		if candidate.Connections < candidate.MaxConnections {
			filter = append(filter, candidate)
		}
	}

	if len(filter) == 0 {
		return
	}

	sort.Slice(filter, func(i, j int) bool {
		return filter[i].Connections <= filter[j].Connections
	})

	best := filter[0]
	_ = s.logger.Log("msg", "discovered peers", "peer_net_addr", best.NetAddr, "peer_address", best.Address)

	go func() {
		if err := s.Node.Connect(best.NetAddr); err != nil {
			_ = s.logger.Log("msg", "failed to connect to best peer", "peer_net_addr", best.NetAddr, "peer_address", best.Address, "err", err)
		}
	}()
}

func (s *Server) churningPeer() {
	currentPeers := s.Node.Peers()
	if len(currentPeers) == 0 {
		return
	}

	randomIndex := rand.Intn(len(currentPeers))
	randomPeer := currentPeers[randomIndex]

	_ = s.logger.Log("msg", "churning peer", "peer_address", randomPeer.ShortString(8))
	s.Node.Disconnect(randomPeer)
}

func (s *Server) logLoop() {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.closeCh:
			return
		case <-ticker.C:
			peers := s.Node.Peers()
			sort.Slice(peers, func(i, j int) bool {
				return peers[i].Lte(peers[j])
			})
			peersAddrStr := make([]string, len(peers))
			for i, peer := range peers {
				peersAddrStr[i] = peer.ShortString(8)
			}
			peerStr := fmt.Sprintf("[%s]", strings.Join(peersAddrStr, ", "))
			connStr := fmt.Sprintf("(%d/%d)", len(peers), s.MaxPeers)

			_ = s.logger.Log("height", s.Chain.GetCurrentHeight(), "state", s.state.Get(), "conn", connStr, "peer", peerStr)
		}
	}
}

func (s *Server) heartbeatLoop() {
	ticker := time.NewTicker(40 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.closeCh:
			return
		case <-ticker.C:
			if err := s.Provider.Heartbeat(s.getOurInfo()); err != nil {
				_ = s.logger.Log("msg", "failed to heartbeat", "err", err)
			}
		}
	}
}

func (s *Server) loop() {
	s.state.Set(Online)
	_ = s.logger.Log("msg", "server is online")

	rawMsgCh := s.Node.ConsumeRawMessage()
	synchronizerMsgCh := s.Synchronizer.OutgoingMessage()
	// apiTxCh := s.ApiServer.ConsumeTransaction()

	for {
		select {
		case <-s.closeCh:
			return
		case m := <-rawMsgCh:
			s.processRawMessage(m)
		case sm := <-synchronizerMsgCh:
			s.processSynchronizerMessage(sm)
			// case tx := <-apiTxCh:
		}
	}
}

func (s *Server) processRawMessage(rm message.RawMessage) {
	msgType := message.MessageType(rm.Payload()[0])
	switch msgType {
	case message.MessageGossipType:
		s.processGossipMessage(rm)
	case message.MessageSyncType:
		s.processSyncMessage(rm)
	case message.MessageConsensusType:
		s.processConsensusMessage(rm)
	}
}

func (s *Server) processGossipMessage(rm message.RawMessage) {
	m, err := s.GossipMessageCodec.Decode(rm.Payload())
	if err != nil {
		return
	}

	gossip := false
	switch t := m.(type) {
	case *block.Transaction:
		if gossip, err = s.processTransaction(rm.From(), t); err != nil {
			switch {
			case errors.Is(err, core.ErrMempoolFull):
				_ = s.logger.Log("msg", "virtual memory pool is fulled", "err", err)
			default:
				_ = s.logger.Log("msg", "failed to process transaction", "err", err)
			}
		}
	case *block.Block:
		if gossip, err = s.processBlock(rm.From(), t); err != nil {
			switch {
			case errors.Is(err, core.ErrBlockKnown):
				break
			case errors.Is(err, core.ErrFutureBlock):
				s.Synchronizer.NotifyChainLagging()
			case errors.Is(err, core.ErrUnknownParent):
				s.Synchronizer.NotifyForkDetected(rm.From(), t)
			default:
				_ = s.logger.Log("msg", "failed to process block", "err", err)
			}
		}
	}

	if err != nil {
		return
	}

	if gossip {
		payload, err := s.GossipMessageCodec.Encode(m)

		if err != nil {
			_ = s.logger.Log("error", "failed to re-encode gossip message", "err", err)
			return
		}

		go func() {
			_ = s.Node.Broadcast(payload)
		}()
	}
}

func (s *Server) processSyncMessage(rm message.RawMessage) {
	sm, err := s.SyncMessageCodec.Decode(rm.Payload())
	if err != nil {
		return
	}

	s.Synchronizer.HandleMessage(rm.From(), sm)
}

func (s *Server) processConsensusMessage(rm message.RawMessage) {
	_, err := s.ConsensusMessageCodec.Decode(rm.Payload())
	if err != nil {
		return
	}
}

func (s *Server) processSynchronizerMessage(sm synchronizer.SynchronizerMessage) {
	payload, err := s.SyncMessageCodec.Encode(sm.SyncMessage())
	if err != nil {
		return
	}

	if !sm.Address().IsZero() {
		go func() {
			_ = s.Node.Send(sm.Address(), payload)
		}()

		return
	}

	go func() {
		_ = s.Node.Broadcast(payload)
	}()
}

func (s *Server) processTransaction(from types.Address, tx *block.Transaction) (bool, error) {
	if !s.state.Eq(Online) {
		return false, nil
	}

	_ = s.logger.Log("msg", "processing transaction", "from", from.ShortString(8))

	if _, err := tx.Hash(); err != nil {
		return false, err
	}

	if err := tx.Verify(); err != nil {
		return false, err
	}

	if s.VirtualMemoryPool.Contains(tx) {
		return false, nil
	}

	tx.FirstSeen()

	if err := s.VirtualMemoryPool.Put(tx); err != nil {
		return false, err
	}

	return true, nil
}

func (s *Server) processBlock(from types.Address, b *block.Block) (bool, error) {
	if !s.state.Eq(Online) {
		return false, nil
	}

	_ = s.logger.Log("msg", "processing block", "from", from.ShortString(8), "height", b.Header.Height)

	if _, err := b.Hash(); err != nil {
		return false, err
	}

	if err := b.Verify(); err != nil {
		return false, err
	}

	if err := s.Chain.AddBlock(b); err != nil {
		return false, err
	}

	s.VirtualMemoryPool.Prune(b.Body.Transactions)

	return true, nil
}
