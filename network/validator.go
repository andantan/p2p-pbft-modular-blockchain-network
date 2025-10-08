package network

import (
	"fmt"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/network/message"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/types"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/util"
	"github.com/go-kit/log"
	"sort"
)

type Validator interface {
	Sign(message.ConsensusMessage) error
	Verify(message.ConsensusMessage) error
	UpdateValidatorSet([]types.Address)
	ProcessConsensusMessage(message.ConsensusMessage) error
}

type PbftValidator struct {
	logger       log.Logger
	privKey      *crypto.PrivateKey
	validatorSet *types.AtomicSet[types.Address]
}

func NewPbftValidator(k *crypto.PrivateKey) *PbftValidator {
	return &PbftValidator{
		logger:       util.LoggerWithPrefixes("Validator"),
		privKey:      k,
		validatorSet: types.NewAtomicSet[types.Address](),
	}
}

func (v *PbftValidator) Sign(msg message.ConsensusMessage) error {
	return msg.Sign(v.privKey)
}

func (v *PbftValidator) Verify(msg message.ConsensusMessage) error {
	addr := msg.Address()
	if !v.validatorSet.Contains(addr) {
		return fmt.Errorf("received %T from non-validator: %s", msg, addr.ShortString(8))
	}

	if err := msg.Verify(); err != nil {
		return err
	}

	return nil
}

func (v *PbftValidator) UpdateValidatorSet(s []types.Address) {
	v.validatorSet.Reset(s)
}

func (v *PbftValidator) ProcessConsensusMessage(msg message.ConsensusMessage) error {
	if err := v.Verify(msg); err != nil {
		return err
	}

	switch m := msg.(type) {
	case *message.PbftPrePrepareMessage:
		return v.processPrePrepare(m)
	case *message.PbftPrepareMessage:
		return v.processPrepare(m)
	case *message.PbftCommitMessage:
		return v.processCommit(m)
	default:
		return fmt.Errorf("unknown consensus message type: %T", m)
	}
}

func (v *PbftValidator) getLeader(view, sequence uint64) types.Address {
	validators := v.validatorSet.Values()

	sort.Slice(validators, func(i, j int) bool {
		return validators[i].Lte(validators[j])
	})

	leaderIndex := int((view + sequence) % uint64(len(validators)))

	return validators[leaderIndex]
}

func (v *PbftValidator) processPrePrepare(msg *message.PbftPrePrepareMessage) error {
	expectedLeader := v.getLeader(msg.View, msg.Sequence)
	if !msg.Address().Equal(expectedLeader) {
		return fmt.Errorf("invalid leader: expected %s, got %s", expectedLeader.ShortString(8), msg.Address().ShortString(8))
	}

	_ = v.logger.Log("msg", "pre-prepare message validated", "address", msg.Address().ShortString(8))
	return nil
}

func (v *PbftValidator) processPrepare(msg *message.PbftPrepareMessage) error {
	_ = v.logger.Log("msg", "prepare message validated", "address", msg.Address().ShortString(8))
	return nil
}

func (v *PbftValidator) processCommit(msg *message.PbftCommitMessage) error {
	_ = v.logger.Log("msg", "commit message validated", "address", msg.Address().ShortString(8))
	return nil
}
