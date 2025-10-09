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
	externalMsgCh    chan network.ConsensusMessage
	internalMsgCh    chan network.ConsensusMessage
	closeCh          chan struct{}
	closeOnce        sync.Once
}

func NewPbftConsensusEngine(
	k *crypto.PrivateKey,
	p core.Processor,
	validatorSet []types.Address,
	finalizedBlockCh chan *block.Block,
	externalMsgCh chan network.ConsensusMessage,
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
		externalMsgCh:     externalMsgCh,
		internalMsgCh:     make(chan network.ConsensusMessage, 100),
		closeCh:           make(chan struct{}),
	}

	return e
}

func (e *PbftConsensusEngine) Start() {
	if !e.state.Eq(Initialized) {
		return
	}

	e.state.Set(Idle)
	_ = e.logger.Log("msg", "changed state", "state", e.state.Get(), "view", e.view.Get(), "sequence", "not-initialized")

	go e.run()
}

func (e *PbftConsensusEngine) HandleMessage(m network.ConsensusMessage) {
	select {
	case <-e.closeCh:
		return
	default:
	}

	select {
	case e.internalMsgCh <- m:
	}
}

func (e *PbftConsensusEngine) Stop() {
	e.closeOnce.Do(func() {
		close(e.closeCh)
		close(e.internalMsgCh)
	})
}

func (e *PbftConsensusEngine) Finalize() {
	select {
	case <-e.closeCh:
		return
	default:
	}

	votes := e.commitVotes.Values()
	commitVotes := make([]*block.CommitVote, len(votes))
	for i, v := range votes {
		commitVotes[i] = block.NewCommitVote(v.Digest, v.PublicKey, v.Signature)
	}

	if err := e.block.Seal(commitVotes, e.validator.GetValidatorSets()); err != nil {
		_ = e.logger.Log("msg", "block_seal_failed", "view", e.view.Get(), "sequence", e.sequence, "err", err)
		// e.startViewChange()
		return
	}

	select {
	case e.finalizedBlockCh <- e.block:
		e.state.Set(Finalized)
		_ = e.logger.Log("msg", "changed state", "state", e.state.Get(), "view", e.view.Get(), "sequence", e.sequence)
	}
}

func (e *PbftConsensusEngine) handlePrePrepareMessage(m *PbftPrePrepareMessage) (network.ConsensusMessage, error) {
	if !e.state.Eq(Idle) {
		return nil, nil
	}

	if err := e.validator.ProcessConsensusMessage(m); err != nil {
		return nil, err
	}

	if err := e.processor.ProcessBlock(m.Block); err != nil {
		if errors.Is(err, core.ErrBlockKnown) {
			return nil, nil
		}

		return nil, err
	}

	e.block = m.Block
	e.sequence = m.Block.Header.Height
	e.state.Set(PrePrepared)
	_ = e.logger.Log("msg", "changed state", "state", e.state.Get(), "view", e.view.Get(), "sequence", e.sequence)

	bh, err := e.block.Hash()
	if err != nil {
		return nil, err
	}

	msg := NewPbftPrepareMessage(m.View, m.Sequence, bh)
	if err = e.validator.Sign(msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func (e *PbftConsensusEngine) handlePrepareMessage(m *PbftPrepareMessage) (network.ConsensusMessage, error) {
	if !e.state.Eq(PrePrepared) {
		return nil, nil
	}

	if err := e.validator.ProcessConsensusMessage(m); err != nil {
		return nil, err
	}

	bh, err := e.block.Hash()
	if err != nil {
		return nil, err
	}

	if !m.BlockHash.Equal(bh) {
		return nil, fmt.Errorf("received prepare vote for incorrect block hash")
	}

	from := m.Address()
	if e.prepareVotes.Exists(from) {
		return nil, nil
	}

	e.prepareVotes.Put(from, m)
	voteCount := e.prepareVotes.Len()
	countLogMsg := fmt.Sprintf("(%d/%d)", voteCount, e.quorum)
	_ = e.logger.Log("msg", "received prepare vote", "address", from.ShortString(8), "vote_count", countLogMsg, "view", e.view.Get(), "sequence", e.sequence)

	e.lock.Lock()
	defer e.lock.Unlock()

	if voteCount >= e.quorum && e.state.Eq(PrePrepared) {
		e.state.Set(Prepared)
		_ = e.logger.Log("msg", "changed state", "state", e.state.Get(), "view", e.view.Get(), "sequence", e.sequence)

		msg := NewPbftCommitMessage(m.View, m.Sequence, bh)
		if err = e.validator.Sign(msg); err != nil {
			return nil, err
		}

		return msg, nil
	}

	return m, nil
}

func (e *PbftConsensusEngine) handleCommitMessage(m *PbftCommitMessage) (network.ConsensusMessage, error) {
	if !e.state.Eq(Prepared) {
		return nil, nil
	}

	if err := e.validator.ProcessConsensusMessage(m); err != nil {
		return nil, err
	}

	bh, err := e.block.Hash()
	if err != nil {
		return nil, err
	}

	if !m.BlockHash.Equal(bh) {
		return nil, fmt.Errorf("received commit vote for incorrect block hash")
	}

	from := m.Address()
	if e.commitVotes.Exists(from) {
		return nil, nil
	}

	e.commitVotes.Put(from, m)
	voteCount := e.commitVotes.Len()
	countLogMsg := fmt.Sprintf("(%d/%d)", voteCount, e.quorum)
	_ = e.logger.Log("msg", "received commit vote", "address", from.ShortString(8), "vote_count", countLogMsg, "view", e.view.Get(), "sequence", e.sequence)

	e.lock.Lock()
	defer e.lock.Unlock()

	if voteCount >= e.quorum && e.state.Eq(Prepared) {
		e.state.Set(Committed)
		_ = e.logger.Log("msg", "changed state", "state", e.state.Get(), "view", e.view.Get(), "sequence", e.sequence)

		e.Finalize()
		return nil, nil
	}

	return m, nil
}

func (e *PbftConsensusEngine) handleViewChangeMessage(m *PbftViewChangeMessage) (network.ConsensusMessage, error) {
	if e.state.Eq(Finalized) {
		return nil, nil
	}

	if err := e.validator.ProcessConsensusMessage(m); err != nil {
		return nil, err
	}

	if m.NewView <= e.view.Get() {
		return nil, nil
	}

	from := m.Address()
	if e.viewChangeVotes.Exists(from) {
		return nil, nil
	}

	e.viewChangeVotes.Put(from, m)
	voteCount := e.viewChangeVotes.Len()
	countLogMsg := fmt.Sprintf("(%d/%d)", voteCount, e.quorum)
	_ = e.logger.Log("msg", "received view-change vote", "address", from.ShortString(8), "vote_count", countLogMsg, "view", e.view.Get(), "sequence", e.sequence)

	newLeader := e.validator.GetLeader(e.sequence, m.NewView)
	ourAddr := e.validator.PublicKey().Address()

	e.lock.Lock()
	defer e.lock.Unlock()

	if ourAddr.Equal(newLeader) && e.viewChangeVotes.Len() >= e.quorum {
		if m.NewView > e.view.Get() {
			e.view.Set(m.NewView)
			_ = e.logger.Log("msg", "elected new leader", "new_view", m.NewView, "sequence", e.sequence)

			msg := NewPbftNewViewMessage(m.NewView, e.sequence, e.block, e.validator.PublicKey(), e.viewChangeVotes.Values())
			if err := e.validator.Sign(msg); err != nil {
				return nil, err
			}

			return msg, nil
		}
	}

	return m, nil
}

func (e *PbftConsensusEngine) handleNewViewMessage(m *PbftNewViewMessage) (network.ConsensusMessage, error) {
	if e.state.Eq(Finalized) {
		return nil, nil
	}

	if err := e.validator.ProcessConsensusMessage(m); err != nil {
		return nil, err
	}

	if m.NewView <= e.view.Get() {
		return nil, nil
	}

	if len(m.ViewChangeMessages) < e.quorum {
		return nil, fmt.Errorf("not enough view change messages in new view message")
	}

	for _, vcm := range m.ViewChangeMessages {
		if err := e.validator.Verify(vcm); err != nil {
			return nil, fmt.Errorf("invalid view change message in new view proof: %w", err)
		}
	}

	_ = e.logger.Log("msg", "accepted new view", "new_view", m.NewView, "sequence", e.sequence)

	e.view.Set(m.NewView)
	e.prepareVotes.Clear()
	e.commitVotes.Clear()
	e.viewChangeVotes.Clear()

	return e.handlePrePrepareMessage(m.PrePrepareMessage)
}

func (e *PbftConsensusEngine) run() {
	timer := time.NewTimer(e.viewChangeTimeout)
	defer timer.Stop()

	_ = e.logger.Log("event", "started consensus engine", "state", e.state.Get(), "view", e.view.Get(), "sequence", "not-initialized")

	for {
		select {
		case <-e.closeCh:
			return
		case <-timer.C:
			_ = e.logger.Log("msg", "consensus timeout", "view", e.view.Get(), "sequence", e.sequence)

			if err := e.startViewChange(); err != nil {
				_ = e.logger.Log("error", "failed to start view change", "err", err, "view", e.view.Get(), "sequence", e.sequence)
			}

			timer.Reset(e.viewChangeTimeout)
		case msg := <-e.internalMsgCh:
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(e.viewChangeTimeout)

			var (
				rm  network.ConsensusMessage
				err error
			)
			switch t := msg.(type) {
			case *PbftPrePrepareMessage:
				rm, err = e.handlePrePrepareMessage(t)
			case *PbftPrepareMessage:
				rm, err = e.handlePrepareMessage(t)
			case *PbftCommitMessage:
				rm, err = e.handleCommitMessage(t)
			case *PbftViewChangeMessage:
				rm, err = e.handleViewChangeMessage(t)
			case *PbftNewViewMessage:
				rm, err = e.handleNewViewMessage(t)
			default:
				rm, err = nil, fmt.Errorf("unknown consensus message type: %T", msg)
			}

			if err != nil {
				errMsg := fmt.Sprintf("failed to handle consensus message type: %T", msg)
				_ = e.logger.Log("msg", errMsg, "err", err)
				continue
			}

			if msg != nil {
				e.send(rm)
			}
		}
	}
}

func (e *PbftConsensusEngine) send(m network.ConsensusMessage) {
	select {
	case <-e.closeCh:
		return
	default:
	}

	select {
	case e.externalMsgCh <- m:
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

	e.send(msg)

	return nil
}
