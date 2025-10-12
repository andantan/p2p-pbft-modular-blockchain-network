package server

import (
	"fmt"
	"github.com/andantan/modular-blockchain/crypto"
	"github.com/andantan/modular-blockchain/network/message"
	"github.com/andantan/modular-blockchain/network/provider"
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
	Syncing
	Online
	Forked
	InConsensus
	Offline
)

func (s ServerState) String() string {
	switch s {
	case Initialized:
		return "Initialized"
	case Bootstrap:
		return "BOOTSTRAP"
	case Syncing:
		return "SYNCING"
	case Online:
		return "ONLINE"
	case Forked:
		return "FORKED"
	case InConsensus:
		return "IN_CONSENSUS"
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
	// apiTxCh := s.ApiServer.ConsumeTransaction()

	for {
		select {
		case <-s.closeCh:
			return
		case msg := <-rawMsgCh:
			if err := s.processMessage(msg); err != nil {
				_ = s.logger.Log("msg", "failed to process message", "err", err)
				continue
			}

			// case tx := <-apiTxCh:
		}
	}
}

func (s *Server) processMessage(rm message.RawMessage) error {
	msgType := message.MessageType(rm.Payload()[0])
	switch msgType {
	case message.MessageGossipType:
		return s.processGossipMessage(rm)
	case message.MessageSyncType:
		return s.processSyncMessage(rm)
	case message.MessageConsensusType:
		return s.processConsensusMessage(rm)
	default:
		return fmt.Errorf("unknown message type: %d", msgType)
	}
}

func (s *Server) processGossipMessage(rm message.RawMessage) error {
	_, err := s.GossipMessageCodec.Decode(rm.Payload())
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) processSyncMessage(rm message.RawMessage) error {
	_, err := s.SyncMessageCodec.Decode(rm.Payload())
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) processConsensusMessage(rm message.RawMessage) error {
	_, err := s.ConsensusMessageCodec.Decode(rm.Payload())
	if err != nil {
		return err
	}

	return nil
}
