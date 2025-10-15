package synchronizer

import (
	"errors"
	"fmt"
	"github.com/andantan/modular-blockchain/core"
	"github.com/andantan/modular-blockchain/core/block"
	"github.com/andantan/modular-blockchain/types"
	"github.com/andantan/modular-blockchain/util"
	"github.com/go-kit/log"
	"math/rand"
	"sort"
	"sync"
	"time"
)

type Synchronizer interface {
	Start()
	HandleMessage(types.Address, SyncMessage)
	OutgoingMessage() <-chan SynchronizerMessage
	IsSynchronized() bool
	NotifyForkDetected(types.Address, *block.Block)
	NotifyChainLagging()
	Stop()
}

type SyncState byte

const (
	Initialized SyncState = iota
	Idle
	Synchronized
	SyncingBlocks
	ResolvingFork
	Terminated
)

func (s SyncState) String() string {
	switch s {
	case Initialized:
		return "Initialized"
	case Idle:
		return "Idle"
	case Synchronized:
		return "Synchronized"
	case SyncingBlocks:
		return "SyncingBlocks"
	case ResolvingFork:
		return "ResolvingFork"
	case Terminated:
		return "Terminated"
	default:
		return "Unknown"
	}
}

type PeerState struct {
	Address   types.Address
	Height    uint64
	State     SyncState
	BlockHash types.Hash
}

type InternalSyncMessage struct {
	From    types.Address
	Message SyncMessage
}

type ChainSynchroizer struct {
	logger log.Logger

	address types.Address
	netAddr string

	state          *types.AtomicNumber[SyncState]
	availablePeers *types.AtomicMap[types.Address, *PeerState]

	chain            core.Chain
	genesisBlockHash types.Hash

	internalMsgCh chan *InternalSyncMessage
	outgoingMsgCh chan SynchronizerMessage
	closeCh       chan struct{}

	closeOnce sync.Once
}

func NewChainSynchronizer(address types.Address, netAddr string, chain core.Chain) *ChainSynchroizer {
	return &ChainSynchroizer{
		logger:         util.LoggerWithPrefixes("Synchronizer"),
		address:        address,
		netAddr:        netAddr,
		state:          types.NewAtomicNumber[SyncState](Initialized),
		availablePeers: types.NewAtomicMap[types.Address, *PeerState](),
		chain:          chain,
		internalMsgCh:  make(chan *InternalSyncMessage, 100),
		outgoingMsgCh:  make(chan SynchronizerMessage, 100),
		closeCh:        make(chan struct{}),
	}
}

func (s *ChainSynchroizer) Start() {
	if !s.state.CompareAndSwap(Initialized, Idle) {
		return
	}

	gh, err := s.chain.GetHeaderByHeight(0)
	if err != nil {
		panic(fmt.Sprintf("Synchronizer error: %s", err))
	}

	gbh, err := gh.Hash()
	if err != nil {
		panic(fmt.Sprintf("Synchronizer error: %s", err))
	}

	s.genesisBlockHash = gbh

	_ = s.logger.Log("msg", "changed state", "state", s.state.Get())

	go s.run()
}

func (s *ChainSynchroizer) HandleMessage(address types.Address, m SyncMessage) {
	if s.state.Gte(Terminated) {
		return
	}

	select {
	case <-s.closeCh:
		return
	default:
	}

	im := &InternalSyncMessage{
		From:    address,
		Message: m,
	}

	select {
	case s.internalMsgCh <- im:
	case <-s.closeCh:
	}
}

func (s *ChainSynchroizer) OutgoingMessage() <-chan SynchronizerMessage {
	return s.outgoingMsgCh
}

func (s *ChainSynchroizer) IsSynchronized() bool {
	return s.state.Eq(Synchronized)
}

func (s *ChainSynchroizer) NotifyForkDetected(from types.Address, b *block.Block) {
	_ = s.logger.Log("msg", "received fork-detect notification", "from", from.ShortString(8), "block_height", b.Header.Height)

	s.state.Set(ResolvingFork)
	_ = s.logger.Log("msg", "changed state", "state", s.state.Get())

	req := &RequestHeadersMessage{
		From:  s.chain.GetCurrentHeight(),
		Count: 100,
	}

	msg := &DirectedMessage{
		To:      from,
		Message: req,
	}

	s.sendToOutgoing(msg)
}

func (s *ChainSynchroizer) NotifyChainLagging() {
	if s.state.Eq(Idle) || s.state.Eq(Synchronized) {
		_ = s.logger.Log("msg", "received chain lagging notification")

		s.state.Set(Idle)
		s.synchronize()
	}
}

func (s *ChainSynchroizer) Stop() {
	s.closeOnce.Do(func() {
		_ = s.logger.Log("msg", "stoping synchronizer engine", "state", s.state.Get())

		close(s.closeCh)
		close(s.internalMsgCh)
		close(s.outgoingMsgCh)

		s.state.Set(Terminated)
		_ = s.logger.Log("msg", "changed state", "state", s.state.Get())
		_ = s.logger.Log("msg", "stopped consensus engine", "state", s.state.Get())
	})
}

func (s *ChainSynchroizer) run() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	_ = s.logger.Log("msg", "start Synchronizer", "state", s.state.Get())

	for {
		select {
		case <-s.closeCh:
			_ = s.logger.Log("msg", "exit synchronizer loop", "state", s.state.Get())
			return
		case <-ticker.C:
			s.synchronize()
		case msg := <-s.internalMsgCh:
			if msg == nil {
				continue
			}

			s.handleSyncMessage(msg)
		}
	}
}

func (s *ChainSynchroizer) handleSyncMessage(m *InternalSyncMessage) {
	var (
		err error
		sm  SynchronizerMessage
	)

	switch t := m.Message.(type) {
	case *RequestStatusMessage:
		sm, err = s.handleRequestStatusMessage(m.From, t)
	case *ResponseStatusMessage:
		sm, err = s.handleResponseStatusMessage(m.From, t)
	case *RequestHeadersMessage:
		sm, err = s.handleRequestHeadersMessage(m.From, t)
	case *ResponseHeadersMessage:
		sm, err = s.handleResponseHeadersMessage(m.From, t)
	case *RequestBlocksMessage:
		sm, err = s.handleRequestBlocksMessage(m.From, t)
	case *ResponseBlocksMessage:
		sm, err = s.handleResponseBlocksMessage(m.From, t)
	default:
		sm, err = nil, fmt.Errorf("unknown sync message (type: %T)", m)
	}

	if err != nil {
		_ = s.logger.Log("msg", err)
		return
	}

	if sm != nil {
		s.sendToOutgoing(sm)
	}
}

func (s *ChainSynchroizer) handleRequestStatusMessage(from types.Address, _ *RequestStatusMessage) (SynchronizerMessage, error) {
	_ = s.logger.Log("msg", "received request status message", "peer_address", from.ShortString(8))

	currentBlock, err := s.chain.GetCurrentBlock()
	if err != nil {
		return nil, err
	}

	cbh, err := currentBlock.Hash()
	if err != nil {
		return nil, err
	}

	res := &ResponseStatusMessage{
		Address:          s.address,
		NetAddr:          s.netAddr,
		Version:          currentBlock.Header.Version,
		Height:           currentBlock.Header.Height,
		State:            byte(s.state.Get()),
		GenesisBlockHash: s.genesisBlockHash,
		CurrentBlockHash: cbh,
	}
	msg := &DirectedMessage{
		To:      from,
		Message: res,
	}

	return msg, nil
}

func (s *ChainSynchroizer) handleResponseStatusMessage(from types.Address, m *ResponseStatusMessage) (SynchronizerMessage, error) {
	_ = s.logger.Log("msg", "received response status message", "peer_address", from.ShortString(8))

	if !s.genesisBlockHash.Equal(m.GenesisBlockHash) {
		return nil, nil
	}

	if s.chain.GetCurrentHeight() <= m.Height {
		nps := &PeerState{
			Address:   from,
			Height:    m.Height,
			State:     SyncState(m.State),
			BlockHash: m.CurrentBlockHash,
		}
		s.availablePeers.Put(from, nps)
		_ = s.logger.Log("msg", "updated sync peer", "peer_address", nps.Address.ShortString(8), "peer_height", m.Height)
	}

	return nil, nil
}

func (s *ChainSynchroizer) handleRequestHeadersMessage(from types.Address, m *RequestHeadersMessage) (SynchronizerMessage, error) {
	_ = s.logger.Log("msg", "received request headers message", "peer_address", from.ShortString(8))

	headers := make([]*block.Header, 0)
	for i := uint64(0); i < m.Count; i++ {
		h, err := s.chain.GetHeaderByHeight(m.From + i)
		if err != nil {
			break
		}
		headers = append(headers, h)
	}

	res := &ResponseHeadersMessage{
		Headers: headers,
	}
	msg := &DirectedMessage{
		To:      from,
		Message: res,
	}

	return msg, nil
}

func (s *ChainSynchroizer) handleResponseHeadersMessage(from types.Address, m *ResponseHeadersMessage) (SynchronizerMessage, error) {
	_ = s.logger.Log("msg", "received response headers message", "peer_address", from.ShortString(8))

	if !s.state.Eq(ResolvingFork) {
		return nil, nil
	}

	if len(m.Headers) == 0 {
		s.state.Set(Idle)
		return nil, nil
	}

	var forkPoint uint64
	forked := false

	for _, h := range m.Headers {
		if !s.chain.HasBlockHash(h.PrevBlockHash) {
			forkPoint = h.Height - 1
			forked = true

			_ = s.logger.Log("msg", "fork point detected", "forked-height", forkPoint)
			break
		}
	}

	if !forked {
		return nil, nil
	}

	if err := s.chain.Rollback(forkPoint); err != nil {
		s.state.Set(Idle)
		return nil, err
	}

	req := &RequestBlocksMessage{
		From:  forkPoint,
		Count: 16,
	}
	msg := &DirectedMessage{
		To:      from,
		Message: req,
	}

	s.state.Set(SyncingBlocks)
	_ = s.logger.Log("msg", "changed state", "state", s.state.Get())

	return msg, nil
}

func (s *ChainSynchroizer) handleRequestBlocksMessage(from types.Address, m *RequestBlocksMessage) (SynchronizerMessage, error) {
	_ = s.logger.Log("msg", "received request blocks message", "peer_address", from.ShortString(8))

	blocks := make([]*block.Block, 0)
	for i := uint64(0); i < m.Count; i++ {
		h, err := s.chain.GetBlockByHeight(m.From + i)
		if err != nil {
			break
		}
		blocks = append(blocks, h)
	}

	res := &ResponseBlocksMessage{
		Blocks: blocks,
	}
	msg := &DirectedMessage{
		To:      from,
		Message: res,
	}

	return msg, nil
}

func (s *ChainSynchroizer) handleResponseBlocksMessage(from types.Address, m *ResponseBlocksMessage) (SynchronizerMessage, error) {
	_ = s.logger.Log("msg", "received response blocks message", "peer_address", from.ShortString(8))

	if !s.state.Eq(SyncingBlocks) {
		return nil, nil
	}

	if len(m.Blocks) == 0 {
		return nil, nil
	}

	sort.Slice(m.Blocks, func(i, j int) bool {
		return m.Blocks[i].Header.Height < m.Blocks[j].Header.Height
	})

	for _, b := range m.Blocks {
		if err := s.chain.AddBlock(b); err != nil {
			if errors.Is(err, core.ErrBlockKnown) {
				continue
			}

			if errors.Is(err, core.ErrFutureBlock) {
				s.NotifyChainLagging()
				return nil, nil
			}

			if errors.Is(err, core.ErrUnknownParent) {
				s.NotifyForkDetected(from, b)
				return nil, nil
			}

			_ = s.logger.Log("msg", "failed to add synced block", "err", err)
			s.state.Set(Idle)
			_ = s.logger.Log("msg", "changed state", "state", s.state.Get())
			return nil, err
		}
	}

	s.state.Set(Idle)
	_ = s.logger.Log("msg", "changed state", "state", s.state.Get())

	return nil, nil
}

func (s *ChainSynchroizer) synchronize() {
	currentState := s.state.Get()
	if currentState != Idle && currentState != Synchronized {
		return
	}

	s.removeBehindPeers()

	bestPeer := s.findBestPeer()

	if bestPeer == nil {
		_ = s.logger.Log("msg", "no available peers to sync with, broadcasting for status")
		m := &BroadcastMessage{
			Message: &RequestStatusMessage{},
		}
		s.sendToOutgoing(m)
		return
	}

	ourHeight := s.chain.GetCurrentHeight()
	if ourHeight < bestPeer.Height {
		_ = s.logger.Log("msg", "chain is behind, requesting blocks", "our_height", ourHeight, "network_height", bestPeer.Height)
		req := &RequestBlocksMessage{
			From:  ourHeight + 1,
			Count: 16,
		}
		m := &DirectedMessage{
			To:      bestPeer.Address,
			Message: req,
		}
		s.sendToOutgoing(m)
		s.state.Set(SyncingBlocks)
		_ = s.logger.Log("msg", "changed state", "state", s.state.Get())
		return
	} else {
		s.state.Set(Synchronized)
		_ = s.logger.Log("msg", "changed state", "state", s.state.Get())
	}

	_ = s.logger.Log("state", s.state.Get(), "available_peers", s.availablePeers.Len())
}

func (s *ChainSynchroizer) findBestPeer() *PeerState {
	highestHeight := uint64(0)
	ourHeight := s.chain.GetCurrentHeight()
	bestPeers := make([]*PeerState, 0)

	s.availablePeers.Range(func(addr types.Address, state *PeerState) bool {
		if state.Height >= ourHeight {
			if state.Height > highestHeight {
				highestHeight = state.Height
				bestPeers = []*PeerState{state}
			} else if state.Height == highestHeight {
				bestPeers = append(bestPeers, state)
			}
		}

		return true
	})

	if len(bestPeers) == 0 {
		return nil
	}

	randomIndex := rand.Intn(len(bestPeers))
	return bestPeers[randomIndex]
}

func (s *ChainSynchroizer) removeBehindPeers() {
	currentHeight := s.chain.GetCurrentHeight()
	thresholdHeight := uint64(float64(currentHeight) * 0.8) // 20%

	behindPeers := make([]*PeerState, 0)
	s.availablePeers.Range(func(addr types.Address, state *PeerState) bool {
		if state.Height < thresholdHeight {
			behindPeers = append(behindPeers, state)
		}
		return true
	})

	for _, peer := range behindPeers {
		s.availablePeers.Remove(peer.Address)
	}
}

func (s *ChainSynchroizer) sendToOutgoing(m SynchronizerMessage) {
	select {
	case <-s.closeCh:
		return
	default:
	}

	select {
	case s.outgoingMsgCh <- m:
	case <-s.closeCh:
		return
	}
}
