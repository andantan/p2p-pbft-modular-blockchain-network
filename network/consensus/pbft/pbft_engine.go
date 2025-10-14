package pbft

import (
	"errors"
	"fmt"
	"github.com/andantan/modular-blockchain/core"
	"github.com/andantan/modular-blockchain/core/block"
	"github.com/andantan/modular-blockchain/crypto"
	"github.com/andantan/modular-blockchain/network/consensus"
	"github.com/andantan/modular-blockchain/types"
	"github.com/andantan/modular-blockchain/util"
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
	Terminated
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
	case Terminated:
		return "Terminated"
	default:
		return "Unknown"
	}
}

type PbftConsensusEngine struct {
	logger log.Logger

	state    *types.AtomicNumber[PbftState]
	block    *block.Block
	view     *types.AtomicNumber[uint64]
	sequence uint64

	processor core.Processor

	validator *PbftValidator
	quorum    int

	prepareVotes      *types.AtomicMap[types.Address, *PbftPrepareMessage]
	commitVotes       *types.AtomicMap[types.Address, *PbftCommitMessage]
	viewChangeVotes   *types.AtomicMap[types.Address, *PbftViewChangeMessage]
	viewChangeTimeout time.Duration

	finalizedBlockCh chan *block.Block
	outgoingMsgCh    chan consensus.ConsensusMessage
	internalMsgCh    chan consensus.ConsensusMessage
	closeCh          chan struct{}

	finalizeOnce sync.Once
	closeOnce    sync.Once
}

func NewPbftConsensusEngine(
	k *crypto.PrivateKey,
	b *block.Block,
	p core.Processor,
	validators []types.Address,
) *PbftConsensusEngine {
	v := NewPbftValidator(k)
	v.UpdateValidatorSet(validators)

	e := &PbftConsensusEngine{
		logger:            util.LoggerWithPrefixes("ConsensusEngine"),
		block:             b,
		state:             types.NewAtomicNumber[PbftState](Initialized),
		view:              types.NewAtomicNumber[uint64](0),
		sequence:          b.Header.Height,
		processor:         p,
		validator:         v,
		quorum:            (2 * len(validators) / 3) + 1,
		prepareVotes:      types.NewAtomicMap[types.Address, *PbftPrepareMessage](),
		commitVotes:       types.NewAtomicMap[types.Address, *PbftCommitMessage](),
		viewChangeVotes:   types.NewAtomicMap[types.Address, *PbftViewChangeMessage](),
		viewChangeTimeout: 5 * time.Second,
		finalizedBlockCh:  make(chan *block.Block, 1),
		outgoingMsgCh:     make(chan consensus.ConsensusMessage, 200),
		internalMsgCh:     make(chan consensus.ConsensusMessage, 200),
		closeCh:           make(chan struct{}),
	}

	return e
}

func (e *PbftConsensusEngine) Start() {
	if !e.state.Eq(Initialized) {
		return
	}

	e.state.Set(Idle)
	_ = e.logger.Log("msg", "changed state", "state", e.state.Get(), "view", e.view.Get(), "sequence", e.sequence)

	go e.run()
}

func (e *PbftConsensusEngine) HandleMessage(m consensus.ConsensusMessage) {
	if e.state.Gte(Finalized) {
		return
	}

	select {
	case <-e.closeCh:
		return
	default:
	}

	select {
	case e.internalMsgCh <- m:
	case <-e.closeCh:
		return
	}
}

func (e *PbftConsensusEngine) OutgoingMessage() <-chan consensus.ConsensusMessage {
	return e.outgoingMsgCh
}

func (e *PbftConsensusEngine) FinalizedBlock() <-chan *block.Block {
	return e.finalizedBlockCh
}

func (e *PbftConsensusEngine) Stop() {
	e.closeOnce.Do(func() {
		e.state.Set(Terminated)
		_ = e.logger.Log("msg", "changed state", "state", e.state.Get(), "view", e.view.Get(), "sequence", e.sequence)
		_ = e.logger.Log("msg", "stoping consensus engine", "state", e.state.Get(), "view", e.view.Get(), "sequence", e.sequence)

		close(e.closeCh)
		close(e.internalMsgCh)
		close(e.outgoingMsgCh)
		close(e.finalizedBlockCh)

		_ = e.logger.Log("msg", "stopped consensus engine", "state", e.state.Get(), "view", e.view.Get(), "sequence", e.sequence)
	})
}

func (e *PbftConsensusEngine) run() {
	timer := time.NewTimer(e.viewChangeTimeout)
	defer timer.Stop()

	_ = e.logger.Log("msg", "start consensus engine", "state", e.state.Get(), "view", e.view.Get(), "sequence", e.sequence)

	for {
		select {
		case <-e.closeCh:
			_ = e.logger.Log("msg", "exit engine loop", "state", e.state.Get(), "view", e.view.Get(), "sequence", e.sequence)
			return
		case <-timer.C:
			_ = e.logger.Log("msg", "consensus timeout", "view", e.view.Get(), "sequence", e.sequence)

			if err := e.startViewChange(); err != nil {
				_ = e.logger.Log("error", "failed to start view change", "err", err, "view", e.view.Get(), "sequence", e.sequence)
			}

			timer.Reset(e.viewChangeTimeout)
		case msg := <-e.internalMsgCh:
			if msg == nil {
				continue
			}

			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(e.viewChangeTimeout)

			e.handleConsensusMessage(msg)
		}
	}
}

func (e *PbftConsensusEngine) handleConsensusMessage(m consensus.ConsensusMessage) {
	var (
		err error
		cm  consensus.ConsensusMessage
	)

	switch t := m.(type) {
	case *PbftPrePrepareMessage:
		cm, err = e.handlePrePrepareMessage(t)
	case *PbftPrepareMessage:
		cm, err = e.handlePrepareMessage(t)
	case *PbftCommitMessage:
		cm, err = e.handleCommitMessage(t)
	case *PbftViewChangeMessage:
		cm, err = e.handleViewChangeMessage(t)
	case *PbftNewViewMessage:
		cm, err = e.handleNewViewMessage(t)
	default:
		cm, err = nil, fmt.Errorf("unknown consensus message (type: %T)", m)
	}

	if err != nil {
		_ = e.logger.Log("msg", err)
		return
	}

	if cm != nil {
		e.sendToOutgoing(cm)
	}
}

func (e *PbftConsensusEngine) handlePrePrepareMessage(m *PbftPrePrepareMessage) (consensus.ConsensusMessage, error) {
	if e.state.Gte(PrePrepared) {
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

	e.sendToOutgoing(m) // gossip

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

func (e *PbftConsensusEngine) handlePrepareMessage(m *PbftPrepareMessage) (consensus.ConsensusMessage, error) {
	if e.state.Gte(Prepared) {
		return nil, nil
	}

	from := m.Address()
	if e.prepareVotes.Exists(from) {
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

	e.prepareVotes.Put(from, m)
	voteCount := e.prepareVotes.Len()
	countLogMsg := fmt.Sprintf("(%d/%d)", voteCount, e.quorum)
	_ = e.logger.Log("msg", "received prepare vote", "address", from.ShortString(8), "vote_count", countLogMsg, "view", e.view.Get(), "sequence", e.sequence)

	e.sendToOutgoing(m) // gossip

	if voteCount >= e.quorum && e.state.Eq(PrePrepared) {
		msg := NewPbftCommitMessage(m.View, m.Sequence, bh)
		if err = e.validator.Sign(msg); err != nil {
			return nil, err
		}

		e.state.Set(Prepared)
		_ = e.logger.Log("msg", "changed state", "state", e.state.Get(), "view", e.view.Get(), "sequence", e.sequence)

		return msg, nil
	}

	return nil, nil
}

func (e *PbftConsensusEngine) handleCommitMessage(m *PbftCommitMessage) (consensus.ConsensusMessage, error) {
	if e.state.Gte(Committed) {
		return nil, nil
	}

	from := m.Address()
	if e.commitVotes.Exists(from) {
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

	e.commitVotes.Put(from, m)
	voteCount := e.commitVotes.Len()
	countLogMsg := fmt.Sprintf("(%d/%d)", voteCount, e.quorum)
	_ = e.logger.Log("msg", "received commit vote", "address", from.ShortString(8), "vote_count", countLogMsg, "view", e.view.Get(), "sequence", e.sequence)

	e.sendToOutgoing(m) // gossip

	if voteCount >= e.quorum && e.state.Eq(Prepared) {
		_ = e.logger.Log("msg", "block committed", "address", e.validator.PublicKey().Address().ShortString(8))
		e.state.Set(Committed)
		_ = e.logger.Log("msg", "changed state", "state", e.state.Get(), "view", e.view.Get(), "sequence", e.sequence)

		e.finalizeEngine()

		return nil, nil
	}

	return nil, nil
}

func (e *PbftConsensusEngine) handleViewChangeMessage(m *PbftViewChangeMessage) (consensus.ConsensusMessage, error) {
	if e.state.Gte(Finalized) {
		return nil, nil
	}

	if e.view.Gte(m.NewView) {
		return nil, nil
	}

	if e.viewChangeVotes.Len() >= e.quorum {
		return nil, nil
	}

	from := m.Address()
	if e.viewChangeVotes.Exists(from) {
		return nil, nil
	}

	if err := e.validator.ProcessConsensusMessage(m); err != nil {
		return nil, err
	}

	e.viewChangeVotes.Put(from, m)
	voteCount := e.viewChangeVotes.Len()
	countLogMsg := fmt.Sprintf("(%d/%d)", voteCount, e.quorum)
	_ = e.logger.Log("msg", "received view-change vote", "address", from.ShortString(8), "vote_count", countLogMsg, "new_view", m.NewView, "sequence", e.sequence)

	e.sendToOutgoing(m) // gossip

	newLeader := e.validator.GetLeader(e.sequence, m.NewView)
	ourAddr := e.validator.PublicKey().Address()

	if ourAddr.Equal(newLeader) && e.viewChangeVotes.Len() >= e.quorum {
		msg := NewPbftNewViewMessage(m.NewView, e.sequence, e.block, e.validator.PublicKey(), e.viewChangeVotes.Values())
		if err := e.validator.Sign(msg); err != nil {
			return nil, err
		}

		_ = e.logger.Log("msg", "elected new leader. broadcast pre-prepare message", "new_view", m.NewView, "sequence", e.sequence)

		return msg, nil
	}

	return nil, nil
}

func (e *PbftConsensusEngine) handleNewViewMessage(m *PbftNewViewMessage) (consensus.ConsensusMessage, error) {
	if e.state.Gte(Finalized) {
		return nil, nil
	}

	if e.view.Gte(m.NewView) {
		return nil, nil
	}

	if err := e.validator.ProcessConsensusMessage(m); err != nil {
		return nil, err
	}

	if len(m.ViewChangeMessages) < e.quorum {
		return nil, fmt.Errorf("not enough view change messages in new view message")
	}

	for _, vcm := range m.ViewChangeMessages {
		if err := e.validator.Verify(vcm); err != nil {
			return nil, fmt.Errorf("invalid view change message in new view proof: %w", err)
		}
	}

	e.sendToOutgoing(m) // gossip

	e.view.Set(m.NewView)
	_ = e.logger.Log("msg", "accepted new view", "new_view", m.NewView, "sequence", e.sequence)
	e.prepareVotes.Clear()
	e.commitVotes.Clear()
	e.viewChangeVotes.Clear()

	return e.handlePrePrepareMessage(m.PrePrepareMessage)
}

func (e *PbftConsensusEngine) finalizeEngine() {
	e.finalizeOnce.Do(func() {
		_ = e.logger.Log("msg", "finalizing consensus", "state", e.state.Get(), "view", e.view.Get(), "sequence", e.sequence)

		select {
		case <-e.closeCh:
			return
		default:
		}

		vs := e.commitVotes.Values()
		cvs := make([]*block.CommitVote, 0, len(vs))
		for _, v := range vs {
			cv := block.NewCommitVote(v.Digest, v.PublicKey, v.Signature)
			cvs = append(cvs, cv)
		}

		valSet := e.validator.GetValidatorSets()
		if err := e.block.Seal(cvs, valSet); err != nil {
			_ = e.logger.Log("msg", "failed block sealing", "view", e.view.Get(), "sequence", e.sequence, "err", err)
			// e.startViewChange()
			return
		}

		_ = e.logger.Log("msg", "successed block sealing", "view", e.view.Get(), "sequence", e.sequence)

		select {
		case <-e.closeCh:
			return
		case e.finalizedBlockCh <- e.block:
			e.state.Set(Finalized)
			_ = e.logger.Log("msg", "changed state", "state", e.state.Get(), "view", e.view.Get(), "sequence", e.sequence)
		}

		_ = e.logger.Log("msg", "finalized consensus", "state", e.state.Get(), "view", e.view.Get(), "sequence", e.sequence)
	})
}

func (e *PbftConsensusEngine) startViewChange() error {
	e.view.Add(1)
	newView := e.view.Get()
	_ = e.logger.Log("msg", "start view change", "new_view", newView, "sequence", e.sequence)

	msg := NewPbftViewChangeMessage(newView, e.sequence)
	if err := e.validator.Sign(msg); err != nil {
		return err
	}

	e.sendToOutgoing(msg)

	return nil
}

func (e *PbftConsensusEngine) sendToOutgoing(m consensus.ConsensusMessage) {
	select {
	case <-e.closeCh:
		return
	case e.outgoingMsgCh <- m:
	}
}
