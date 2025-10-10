package pbft

import (
	"fmt"
	"github.com/andantan/modular-blockchain/crypto"
	"github.com/andantan/modular-blockchain/network/message"
	"github.com/andantan/modular-blockchain/types"
	"github.com/andantan/modular-blockchain/util"
	"github.com/go-kit/log"
	"sort"
)

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

func (v *PbftValidator) PublicKey() *crypto.PublicKey {
	return v.privKey.PublicKey()
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
	case *PbftPrePrepareMessage:
		return v.validatePrePrepareMessage(m)
	case *PbftPrepareMessage:
		return v.validatePrepareMessage(m)
	case *PbftCommitMessage:
		return v.validateCommitMessage(m)
	case *PbftViewChangeMessage:
		return v.validateViewChangeMessage(m)
	case *PbftNewViewMessage:
		return v.validateNewViewMessage(m)
	default:
		return fmt.Errorf("unknown consensus message type: %T", m)
	}
}

func (v *PbftValidator) GetValidatorSets() []types.Address {
	return v.validatorSet.Values()
}

func (v *PbftValidator) GetLeader(view, sequence uint64) types.Address {
	validators := v.validatorSet.Values()

	sort.Slice(validators, func(i, j int) bool {
		return validators[i].Lte(validators[j])
	})

	leaderIndex := int((view + sequence) % uint64(len(validators)))

	return validators[leaderIndex]
}

func (v *PbftValidator) validatePrePrepareMessage(m *PbftPrePrepareMessage) error {
	expectedLeader := v.GetLeader(m.View, m.Sequence)
	if !m.Address().Equal(expectedLeader) {
		return fmt.Errorf("invalid leader: expected %s, got %s", expectedLeader.ShortString(8), m.Address().ShortString(8))
	}

	_ = v.logger.Log("msg", "pre-prepare message validated", "address", m.Address().ShortString(8))
	return nil
}

func (v *PbftValidator) validatePrepareMessage(m *PbftPrepareMessage) error {
	_ = v.logger.Log("msg", "prepare message validated", "address", m.Address().ShortString(8))
	return nil
}

func (v *PbftValidator) validateCommitMessage(m *PbftCommitMessage) error {
	_ = v.logger.Log("msg", "commit message validated", "address", m.Address().ShortString(8))
	return nil
}

func (v *PbftValidator) validateViewChangeMessage(m *PbftViewChangeMessage) error {
	_ = v.logger.Log("msg", "view-change message validated", "address", m.Address().ShortString(8))
	return nil
}

func (v *PbftValidator) validateNewViewMessage(m *PbftNewViewMessage) error {
	expectedLeader := v.GetLeader(m.NewView, m.Sequence)
	if !m.Address().Equal(expectedLeader) {
		return fmt.Errorf("invalid new-view leader:  expected %s, got %s", expectedLeader.ShortString(8), m.Address().ShortString(8))
	}

	_ = v.logger.Log("msg", "new-view message validated", "address", m.Address().ShortString(8))
	return nil
}
