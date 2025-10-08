package pbft

import (
	"errors"
	"fmt"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/core"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/core/block"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/network"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/types"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/util"
	"github.com/go-kit/log"
	"sync"
	"time"
)

type PbftState byte

const (
	Initialized PbftState = iota
	Idle
	PrePrepared
	Prepared
	Committed
	Finalized
)

func (s PbftState) String() string {
	switch s {
	case Initialized:
		return "Initialized"
	case Idle:
		return "Idle"
	case PrePrepared:
		return "PrePrepared"
	case Prepared:
		return "Prepared"
	case Committed:
		return "Committed"
	case Finalized:
		return "Finalized"
	default:
		return "Unknown"
	}
}

type PbftConsensusEngine struct {
	logger log.Logger
	lock   sync.RWMutex

	state    *types.AtomicNumber[PbftState]
	view     *types.AtomicNumber[uint64]
	sequence uint64

	block     *block.Block
	processor core.Processor

	validator *PbftValidator
	quorum    int

	prepareVotes      *types.AtomicMap[types.Address, *PbftPrepareMessage]
	commitVotes       *types.AtomicMap[types.Address, *PbftCommitMessage]
	viewChangeVotes   *types.AtomicMap[types.Address, *PbftViewChangeMessage]
	viewChangeTimeout time.Duration

	finalizedBlockCh chan *block.Block
	consensusMsgCh   chan network.ConsensusMessage
	signalCh         chan struct{}
	closeCh          chan struct{}
	closeOnce        sync.Once
}

func NewPbftConsensusEngine(
	k *crypto.PrivateKey,
	p core.Processor,
	validatorSet []types.Address,
	finalizedBlockCh chan *block.Block,
	consensusMsgCh chan network.ConsensusMessage,
) *PbftConsensusEngine {
	v := NewPbftValidator(k)
	v.UpdateValidatorSet(validatorSet)

	e := &PbftConsensusEngine{
		logger:            util.LoggerWithPrefixes("ConsensusEngine"),
		state:             types.NewAtomicNumber[PbftState](Initialized),
		view:              types.NewAtomicNumber[uint64](0),
		processor:         p,
		validator:         v,
		quorum:            (2 * len(validatorSet) / 3) + 1,
		prepareVotes:      types.NewAtomicMap[types.Address, *PbftPrepareMessage](),
		commitVotes:       types.NewAtomicMap[types.Address, *PbftCommitMessage](),
		viewChangeVotes:   types.NewAtomicMap[types.Address, *PbftViewChangeMessage](),
		viewChangeTimeout: 5 * time.Second,
		finalizedBlockCh:  finalizedBlockCh,
		consensusMsgCh:    consensusMsgCh,
		signalCh:          make(chan struct{}),
		closeCh:           make(chan struct{}),
	}

	return e
}

func (e *PbftConsensusEngine) Start() {
	if !e.state.Eq(Initialized) {
		return
	}

	e.state.Set(Idle)
	_ = e.logger.Log("msg", "changed state", "state", e.state, "view", e.view.Get(), "sequence", "not-initialized")

	go e.runTimer()

	_ = e.logger.Log("event", "started consensus engine", "state", e.state, "view", e.view.Get(), "sequence", "not-initialized")
}

func (e *PbftConsensusEngine) HandleMessage(m network.ConsensusMessage) error {
	e.signalProgress()

	switch t := m.(type) {
	case *PbftPrePrepareMessage:
		return e.handlePrePrepareMessage(t)
	case *PbftPrepareMessage:
		return e.handlePrepareMessage(t)
	case *PbftCommitMessage:
		return e.handleCommitMessage(t)
	case *PbftViewChangeMessage:
		return e.handleViewChangeMessage(t)
	case *PbftNewViewMessage:
		return e.handleNewViewMessage(t)
	default:
		return fmt.Errorf("unknown consensus message type: %T", m)
	}
}

func (e *PbftConsensusEngine) Stop() {
	e.closeOnce.Do(func() {
		close(e.closeCh)
		close(e.signalCh)
	})
}

func (e *PbftConsensusEngine) handlePrePrepareMessage(m *PbftPrePrepareMessage) error {
	if !e.state.Eq(Idle) {
		return nil
	}

	if err := e.validator.ProcessConsensusMessage(m); err != nil {
		return err
	}

	if err := e.processor.ProcessBlock(m.Block); err != nil {
		if errors.Is(err, core.ErrBlockKnown) {
			return nil
		}

		return err
	}

	e.block = m.Block
	e.sequence = m.Block.Header.Height
	e.state.Set(PrePrepared)
	_ = e.logger.Log("msg", "changed state", "state", e.state, "view", e.view.Get(), "sequence", e.sequence)

	bh, err := e.block.Hash()
	if err != nil {
		return err
	}

	msg := NewPbftPrepareMessage(m.View, m.Sequence, bh)
	if err = e.validator.Sign(msg); err != nil {
		return err
	}

	return e.sendToConsensusMsgChannel(msg)
}

func (e *PbftConsensusEngine) handlePrepareMessage(m *PbftPrepareMessage) error {
	if !e.state.Eq(PrePrepared) {
		return nil
	}

	if err := e.validator.ProcessConsensusMessage(m); err != nil {
		return err
	}

	bh, err := e.block.Hash()
	if err != nil {
		return err
	}

	if !m.BlockHash.Equal(bh) {
		return fmt.Errorf("received prepare vote for incorrect block hash")
	}

	from := m.Address()
	if e.prepareVotes.Exists(from) {
		return nil
	}

	e.prepareVotes.Put(from, m)
	voteCount := e.prepareVotes.Len()
	countLogMsg := fmt.Sprintf("(%d/%d)", voteCount, e.quorum)
	_ = e.logger.Log("msg", "received prepare vote", "address", from.ShortString(8), "vote_count", countLogMsg)

	e.lock.Lock()
	defer e.lock.Unlock()

	if voteCount >= e.quorum && e.state.Eq(PrePrepared) {
		e.state.Set(Prepared)
		_ = e.logger.Log("msg", "changed state", "state", e.state, "view", e.view.Get(), "sequence", e.sequence)

		msg := NewPbftCommitMessage(m.View, m.Sequence, bh)
		if err = e.validator.Sign(msg); err != nil {
			return err
		}

		return e.sendToConsensusMsgChannel(msg)
	}

	return nil
}

func (e *PbftConsensusEngine) handleCommitMessage(m *PbftCommitMessage) error {
	if !e.state.Eq(Prepared) {
		return nil
	}

	if err := e.validator.ProcessConsensusMessage(m); err != nil {
		return err
	}

	bh, err := e.block.Hash()
	if err != nil {
		return err
	}

	if !m.BlockHash.Equal(bh) {
		return fmt.Errorf("received commit vote for incorrect block hash")
	}

	from := m.Address()
	if e.commitVotes.Exists(from) {
		return nil
	}

	e.commitVotes.Put(from, m)
	voteCount := e.commitVotes.Len()
	countLogMsg := fmt.Sprintf("(%d/%d)", voteCount, e.quorum)
	_ = e.logger.Log("msg", "received commit vote", "address", from.ShortString(8), "vote_count", countLogMsg)

	e.lock.Lock()
	defer e.lock.Unlock()

	if voteCount >= e.quorum && e.state.Eq(Prepared) {
		e.state.Set(Committed)
		_ = e.logger.Log("msg", "changed state", "state", e.state, "view", e.view.Get(), "sequence", e.sequence)

		e.finalizeBlock()
	}

	return nil
}

func (e *PbftConsensusEngine) handleViewChangeMessage(m *PbftViewChangeMessage) error {
	if e.state.Eq(Finalized) {
		return nil
	}

	if err := e.validator.ProcessConsensusMessage(m); err != nil {
		return err
	}

	if m.NewView <= e.view.Get() {
		return nil
	}

	from := m.Address()
	if e.viewChangeVotes.Exists(from) {
		return nil
	}

	e.viewChangeVotes.Put(from, m)
	voteCount := e.viewChangeVotes.Len()
	countLogMsg := fmt.Sprintf("(%d/%d)", voteCount, e.quorum)
	_ = e.logger.Log("msg", "received view-change vote", "address", from.ShortString(8), "vote_count", countLogMsg)

	newLeader := e.validator.GetLeader(e.sequence, m.NewView)
	ourAddr := e.validator.PublicKey().Address()

	e.lock.Lock()
	defer e.lock.Unlock()

	if ourAddr.Equal(newLeader) && e.viewChangeVotes.Len() >= e.quorum {
		if m.NewView > e.view.Get() {
			e.view.Set(m.NewView)
			_ = e.logger.Log("msg", "elected new leader", "new_view", m.NewView, "sequence", e.sequence)

			newViewMsg := NewPbftNewViewMessage(m.NewView, e.sequence, e.block, e.validator.PublicKey(), e.viewChangeVotes.Values())
			if err := e.validator.Sign(newViewMsg); err != nil {
				return err
			}

			return e.sendToConsensusMsgChannel(newViewMsg)
		}
	}

	return nil
}

func (e *PbftConsensusEngine) handleNewViewMessage(m *PbftNewViewMessage) error {
	if e.state.Eq(Finalized) {
		return nil
	}

	if err := e.validator.ProcessConsensusMessage(m); err != nil {
		return err
	}

	if m.NewView <= e.view.Get() {
		return nil
	}

	if len(m.ViewChangeMessages) < e.quorum {
		return fmt.Errorf("not enough view change messages in new view message")
	}

	for _, vcm := range m.ViewChangeMessages {
		if err := e.validator.Verify(vcm); err != nil {
			return fmt.Errorf("invalid view change message in new view proof: %w", err)
		}
	}

	_ = e.logger.Log("msg", "accepted new view", "new_view", m.NewView, "sequence", e.sequence)

	e.view.Set(m.NewView)
	e.prepareVotes.Clear()
	e.commitVotes.Clear()
	e.viewChangeVotes.Clear()

	return e.HandleMessage(m.PrePrepareMessage)
}

func (e *PbftConsensusEngine) sendToConsensusMsgChannel(m network.ConsensusMessage) error {
	select {
	case <-e.closeCh:
		return errors.New("engine is shutting down, view change aborted")
	default:
		e.consensusMsgCh <- m
	}

	return nil
}

func (e *PbftConsensusEngine) finalizeBlock() {
	select {
	case <-e.closeCh:
		return
	case e.finalizedBlockCh <- e.block:
		e.state.Set(Finalized)
		_ = e.logger.Log("msg", "changed state", "state", e.state, "view", e.view.Get(), "sequence", e.sequence)
	}
}

func (e *PbftConsensusEngine) signalProgress() {
	select {
	case <-e.closeCh:
		return
	default:
	}

	select {
	case e.signalCh <- struct{}{}:
	default:
	}
}

func (e *PbftConsensusEngine) runTimer() {
	timer := time.NewTimer(e.viewChangeTimeout)
	defer timer.Stop()

	for {
		select {
		case <-e.closeCh:
			return
		case <-timer.C:
			_ = e.logger.Log("msg", "consensus timeout", "view", e.view.Get(), "sequence", e.sequence)

			if err := e.startViewChange(); err != nil {
				_ = e.logger.Log("error", "failed to start view change", "err", err)
			}

			timer.Reset(e.viewChangeTimeout)
		case <-e.signalCh:
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(e.viewChangeTimeout)
		}
	}
}

func (e *PbftConsensusEngine) startViewChange() error {
	e.view.Add(1)
	newView := e.view.Get()
	_ = e.logger.Log("msg", "start view change", "new_view", newView, "sequence", e.sequence)

	msg := NewPbftViewChangeMessage(newView, e.sequence)
	if err := e.validator.Sign(msg); err != nil {
		return err
	}

	return e.sendToConsensusMsgChannel(msg)
}
