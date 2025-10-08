package pbft

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/core/block"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
	pbBlock "github.com/andantan/p2p-pbft-modular-blockchain-network/proto/core/block"
	pbPbft "github.com/andantan/p2p-pbft-modular-blockchain-network/proto/network/consensus/pbft"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/types"
	"google.golang.org/protobuf/proto"
)

/*
	Phase 1: Pre-Prepare-Request <PRE-PREPARE, v, n, d, m>

Message = <PRE-PREPARE, View, Sequence, <PublicKey, Signature>, Block>
*/
type PbftPrePrepareMessage struct {
	View      uint64
	Sequence  uint64
	Block     *block.Block
	PublicKey *crypto.PublicKey
	Signature *crypto.Signature
}

func NewPbftPrePrepareMessage(view, sequence uint64, b *block.Block, k *crypto.PublicKey) *PbftPrePrepareMessage {
	return &PbftPrePrepareMessage{
		View:      view,
		Sequence:  sequence,
		Block:     b,
		PublicKey: k,
	}
}

func (m *PbftPrePrepareMessage) Hash() (types.Hash, error) {
	if m.Block == nil {
		return types.ZeroHash, fmt.Errorf("cannot hash identity with nil block")
	}

	if m.PublicKey == nil {
		return types.ZeroHash, fmt.Errorf("cannot hash identity with nil public key")
	}

	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.LittleEndian, m.View); err != nil {
		return types.ZeroHash, err
	}
	if err := binary.Write(buf, binary.LittleEndian, m.Sequence); err != nil {
		return types.ZeroHash, err
	}

	blockHash, err := m.Block.Hash()
	if err != nil {
		return types.ZeroHash, err
	}

	buf.Write(blockHash.Bytes())
	buf.Write(m.PublicKey.Bytes())

	hash := sha256.Sum256(buf.Bytes())

	return hash, nil
}

func (m *PbftPrePrepareMessage) Sign(privKey *crypto.PrivateKey) error {
	pubKeyFromPriv := privKey.PublicKey()

	if !bytes.Equal(m.PublicKey.Bytes(), pubKeyFromPriv.Bytes()) {
		return fmt.Errorf("public key in message does not match the private key for signing")
	}

	h, err := m.Hash()
	if err != nil {
		return err
	}

	sig, err := privKey.Sign(h.Bytes())
	if err != nil {
		return err
	}

	m.Signature = sig
	return nil
}

func (m *PbftPrePrepareMessage) Verify() error {
	if m.Signature == nil {
		return fmt.Errorf("PbftPrePrepareMessage has no signature to verify")
	}
	if m.PublicKey == nil {
		return fmt.Errorf("PbftPrePrepareMessage has no public key to verify with")
	}

	h, err := m.Hash()
	if err != nil {
		return err
	}

	if !m.Signature.Verify(m.PublicKey, h.Bytes()) {
		return fmt.Errorf("PbftPrePrepareMessage has invalid signature")
	}

	return nil
}

func (m *PbftPrePrepareMessage) ToProto() (proto.Message, error) {
	bp, err := m.Block.ToProto()
	if err != nil {
		return nil, err
	}

	return &pbPbft.PbftPrePrepareMessage{
		View:      m.View,
		Sequence:  m.Sequence,
		Block:     bp.(*pbBlock.Block),
		PublicKey: m.PublicKey.Bytes(),
		Signature: m.Signature.Bytes(),
	}, nil
}

func (m *PbftPrePrepareMessage) FromProto(msg proto.Message) error {
	p, ok := msg.(*pbPbft.PbftPrePrepareMessage)
	if !ok {
		return errors.New("invalid proto message type for PbftPrePrepareMessage")
	}

	var (
		err error
		b   = new(block.Block)
		key *crypto.PublicKey
		sig *crypto.Signature
	)

	if err = b.FromProto(p.Block); err != nil {
		return err
	}

	if key, err = crypto.PublicKeyFromBytes(p.PublicKey); err != nil {
		return err
	}

	if sig, err = crypto.SignatureFromBytes(p.Signature); err != nil {
		return err
	}

	m.View = p.View
	m.Sequence = p.Sequence
	m.Block = b
	m.PublicKey = key
	m.Signature = sig

	return nil
}

func (m *PbftPrePrepareMessage) EmptyProto() proto.Message {
	return &pbPbft.PbftPrePrepareMessage{}
}

func (m *PbftPrePrepareMessage) Address() types.Address {
	return m.PublicKey.Address()
}

/*
	Phase 2: Prepare-Vote <PREPARE, v, n, d, md>

Message = <PREPARE, View, Sequence, <PublicKey, Signature>, Block>
*/
type PbftPrepareMessage struct {
	View      uint64
	Sequence  uint64
	BlockHash types.Hash
	PublicKey *crypto.PublicKey
	Signature *crypto.Signature
}

func NewPbftPrepareMessage(view, sequence uint64, h types.Hash) *PbftPrepareMessage {
	return &PbftPrepareMessage{
		View:      view,
		Sequence:  sequence,
		BlockHash: h,
	}
}

func (m *PbftPrepareMessage) Hash() (types.Hash, error) {
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.LittleEndian, m.View); err != nil {
		return types.ZeroHash, err
	}
	if err := binary.Write(buf, binary.LittleEndian, m.Sequence); err != nil {
		return types.ZeroHash, err
	}

	buf.Write(m.BlockHash.Bytes())

	hash := sha256.Sum256(buf.Bytes())

	return hash, nil
}

func (m *PbftPrepareMessage) Sign(privKey *crypto.PrivateKey) error {
	h, err := m.Hash()
	if err != nil {
		return err
	}

	sig, err := privKey.Sign(h.Bytes())
	if err != nil {
		return err
	}

	m.PublicKey = privKey.PublicKey()
	m.Signature = sig

	return nil
}

func (m *PbftPrepareMessage) Verify() error {
	if m.Signature == nil {
		return fmt.Errorf("PbftPrepareMessage has no signature to verify")
	}
	if m.PublicKey == nil {
		return fmt.Errorf("PbftPrepareMessage has no public key to verify with")
	}

	h, err := m.Hash()
	if err != nil {
		return err
	}

	if !m.Signature.Verify(m.PublicKey, h.Bytes()) {
		return fmt.Errorf("PbftPrepareMessage has invalid signature")
	}

	return nil
}

func (m *PbftPrepareMessage) ToProto() (proto.Message, error) {
	return &pbPbft.PbftPrepareMessage{
		View:      m.View,
		Sequence:  m.Sequence,
		BlockHash: m.BlockHash.Bytes(),
		PublicKey: m.PublicKey.Bytes(),
		Signature: m.Signature.Bytes(),
	}, nil
}

func (m *PbftPrepareMessage) FromProto(msg proto.Message) error {
	p, ok := msg.(*pbPbft.PbftPrepareMessage)
	if !ok {
		return errors.New("invalid proto message type for PbftPrepareMessage")
	}

	var (
		err error
		h   types.Hash
		key *crypto.PublicKey
		sig *crypto.Signature
	)

	if h, err = types.HashFromBytes(p.BlockHash); err != nil {
		return err
	}

	if key, err = crypto.PublicKeyFromBytes(p.PublicKey); err != nil {
		return err
	}

	if sig, err = crypto.SignatureFromBytes(p.Signature); err != nil {
		return err
	}

	m.View = p.View
	m.Sequence = p.Sequence
	m.BlockHash = h
	m.PublicKey = key
	m.Signature = sig

	return nil
}

func (m *PbftPrepareMessage) EmptyProto() proto.Message {
	return &pbPbft.PbftPrepareMessage{}
}

func (m *PbftPrepareMessage) Address() types.Address {
	return m.PublicKey.Address()
}

/*
	Phase 3: Commit-Vote <COMMIT, v, n, d, md>

Message = <COMMIT, View, Sequence, <PublicKey, Signature>, Block>
*/
type PbftCommitMessage struct {
	View      uint64
	Sequence  uint64
	BlockHash types.Hash
	PublicKey *crypto.PublicKey
	Signature *crypto.Signature
}

func NewPbftCommitMessage(view, sequence uint64, h types.Hash) *PbftCommitMessage {
	return &PbftCommitMessage{
		View:      view,
		Sequence:  sequence,
		BlockHash: h,
	}
}

func (m *PbftCommitMessage) Hash() (types.Hash, error) {
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.LittleEndian, m.View); err != nil {
		return types.ZeroHash, err
	}
	if err := binary.Write(buf, binary.LittleEndian, m.Sequence); err != nil {
		return types.ZeroHash, err
	}

	buf.Write(m.BlockHash.Bytes())

	hash := sha256.Sum256(buf.Bytes())
	return hash, nil
}

func (m *PbftCommitMessage) Sign(privKey *crypto.PrivateKey) error {
	h, err := m.Hash()
	if err != nil {
		return err
	}

	sig, err := privKey.Sign(h.Bytes())
	if err != nil {
		return err
	}

	m.PublicKey = privKey.PublicKey()
	m.Signature = sig

	return nil
}

func (m *PbftCommitMessage) Verify() error {
	if m.Signature == nil {
		return fmt.Errorf("PbftCommitMessage has no signature to verify")
	}
	if m.PublicKey == nil {
		return fmt.Errorf("PbftCommitMessage has no public key to verify with")
	}

	h, err := m.Hash()
	if err != nil {
		return err
	}

	if !m.Signature.Verify(m.PublicKey, h.Bytes()) {
		return fmt.Errorf("PbftCommitMessage has invalid signature")
	}

	return nil
}

func (m *PbftCommitMessage) ToProto() (proto.Message, error) {
	return &pbPbft.PbftCommitMessage{
		View:      m.View,
		Sequence:  m.Sequence,
		BlockHash: m.BlockHash.Bytes(),
		PublicKey: m.PublicKey.Bytes(),
		Signature: m.Signature.Bytes(),
	}, nil
}

func (m *PbftCommitMessage) FromProto(msg proto.Message) error {
	p, ok := msg.(*pbPbft.PbftCommitMessage)
	if !ok {
		return errors.New("invalid proto message type for PbftCommitMessage")
	}

	var (
		err error
		h   types.Hash
		key *crypto.PublicKey
		sig *crypto.Signature
	)

	if h, err = types.HashFromBytes(p.BlockHash); err != nil {
		return err
	}

	if key, err = crypto.PublicKeyFromBytes(p.PublicKey); err != nil {
		return err
	}

	if sig, err = crypto.SignatureFromBytes(p.Signature); err != nil {
		return err
	}

	m.View = p.View
	m.Sequence = p.Sequence
	m.BlockHash = h
	m.PublicKey = key
	m.Signature = sig

	return nil
}

func (m *PbftCommitMessage) EmptyProto() proto.Message {
	return &pbPbft.PbftCommitMessage{}
}

func (m *PbftCommitMessage) Address() types.Address {
	return m.PublicKey.Address()
}

type PbftViewChangeMessage struct {
	NewView   uint64
	Sequence  uint64
	PublicKey *crypto.PublicKey
	Signature *crypto.Signature
}

func NewPbftViewChangeMessage(newView, sequence uint64) *PbftViewChangeMessage {
	return &PbftViewChangeMessage{
		NewView:  newView,
		Sequence: sequence,
	}
}

func (m *PbftViewChangeMessage) Hash() (types.Hash, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, m.NewView); err != nil {
		return types.ZeroHash, err
	}
	if err := binary.Write(buf, binary.LittleEndian, m.Sequence); err != nil {
		return types.ZeroHash, err
	}

	hash := sha256.Sum256(buf.Bytes())
	return hash, nil
}

func (m *PbftViewChangeMessage) Sign(privKey *crypto.PrivateKey) error {
	hash, err := m.Hash()
	if err != nil {
		return err
	}

	sig, err := privKey.Sign(hash.Bytes())
	if err != nil {
		return err
	}

	m.PublicKey = privKey.PublicKey()
	m.Signature = sig
	return nil
}

func (m *PbftViewChangeMessage) Verify() error {
	if m.Signature == nil {
		return fmt.Errorf("PbftViewChangeMessage has no signature to verify")
	}
	if m.PublicKey == nil {
		return fmt.Errorf("PbftViewChangeMessage has no public key to verify with")
	}

	hash, err := m.Hash()
	if err != nil {
		return err
	}

	if !m.Signature.Verify(m.PublicKey, hash.Bytes()) {
		return fmt.Errorf("PbftViewChangeMessage has invalid signature")
	}

	return nil
}

func (m *PbftViewChangeMessage) ToProto() (proto.Message, error) {
	return &pbPbft.PbftViewChangeMessage{
		NewView:   m.NewView,
		Sequence:  m.Sequence,
		PublicKey: m.PublicKey.Bytes(),
		Signature: m.Signature.Bytes(),
	}, nil
}

func (m *PbftViewChangeMessage) FromProto(msg proto.Message) error {
	p, ok := msg.(*pbPbft.PbftViewChangeMessage)
	if !ok {
		return fmt.Errorf("invalid proto message type for PbftViewChangeMessage")
	}

	var (
		err error
		key *crypto.PublicKey
		sig *crypto.Signature
	)

	if key, err = crypto.PublicKeyFromBytes(p.PublicKey); err != nil {
		return err
	}

	if sig, err = crypto.SignatureFromBytes(p.Signature); err != nil {
		return err
	}

	m.NewView = p.NewView
	m.Sequence = p.Sequence
	m.PublicKey = key
	m.Signature = sig

	return nil
}

func (m *PbftViewChangeMessage) EmptyProto() proto.Message {
	return &pbPbft.PbftViewChangeMessage{}
}

func (m *PbftViewChangeMessage) Address() types.Address {
	return m.PublicKey.Address()
}

type PbftNewViewMessage struct {
	NewView            uint64
	Sequence           uint64
	ViewChangeMessages []*PbftViewChangeMessage
	PrePrepareMessage  *PbftPrePrepareMessage
	PublicKey          *crypto.PublicKey
	Signature          *crypto.Signature
}

func NewPbftNewViewMessage(
	view, sequence uint64,
	b *block.Block,
	k *crypto.PublicKey,
	viewChanges []*PbftViewChangeMessage,
) *PbftNewViewMessage {
	return &PbftNewViewMessage{
		NewView:            view,
		Sequence:           sequence,
		PrePrepareMessage:  NewPbftPrePrepareMessage(view, sequence, b, k),
		ViewChangeMessages: viewChanges,
	}
}

func (m *PbftNewViewMessage) Hash() (types.Hash, error) {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, m.NewView)
	_ = binary.Write(buf, binary.LittleEndian, m.Sequence)

	prePrepareHash, err := m.PrePrepareMessage.Hash()
	if err != nil {
		return types.ZeroHash, err
	}
	buf.Write(prePrepareHash.Bytes())

	for _, vcMsg := range m.ViewChangeMessages {
		vcHash, err := vcMsg.Hash()
		if err != nil {
			return types.ZeroHash, err
		}
		buf.Write(vcHash.Bytes())
	}

	hash := sha256.Sum256(buf.Bytes())

	return hash, nil
}

func (m *PbftNewViewMessage) Sign(privKey *crypto.PrivateKey) error {
	hash, err := m.Hash()
	if err != nil {
		return err
	}

	sig, err := privKey.Sign(hash.Bytes())
	if err != nil {
		return err
	}

	m.PublicKey = privKey.PublicKey()
	m.Signature = sig
	return nil
}

func (m *PbftNewViewMessage) Verify() error {
	if m.Signature == nil {
		return fmt.Errorf("PbftNewViewMessage has no signature to verify")
	}
	if m.PublicKey == nil {
		return fmt.Errorf("PbftNewViewMessage has no public key to verify with")
	}

	hash, err := m.Hash()
	if err != nil {
		return err
	}

	if !m.Signature.Verify(m.PublicKey, hash.Bytes()) {
		return fmt.Errorf("PbftNewViewMessage has invalid signature")
	}

	return nil
}

func (m *PbftNewViewMessage) ToProto() (proto.Message, error) {
	prePrepareProto, err := m.PrePrepareMessage.ToProto()
	if err != nil {
		return nil, err
	}

	viewChangesProto := make([]*pbPbft.PbftViewChangeMessage, len(m.ViewChangeMessages))
	for i, vcMsg := range m.ViewChangeMessages {
		vcProto, err := vcMsg.ToProto()
		if err != nil {
			return nil, err
		}
		viewChangesProto[i] = vcProto.(*pbPbft.PbftViewChangeMessage)
	}

	return &pbPbft.PbftNewViewMessage{
		NewView:            m.NewView,
		Sequence:           m.Sequence,
		ViewChangeMessages: viewChangesProto,
		PrePrepareMessage:  prePrepareProto.(*pbPbft.PbftPrePrepareMessage),
		PublicKey:          m.PublicKey.Bytes(),
		Signature:          m.Signature.Bytes(),
	}, nil
}

func (m *PbftNewViewMessage) FromProto(msg proto.Message) error {
	p, ok := msg.(*pbPbft.PbftNewViewMessage)
	if !ok {
		return fmt.Errorf("invalid proto message type for PbftNewViewMessage")
	}

	var (
		err error
		key *crypto.PublicKey
		sig *crypto.Signature
	)

	if key, err = crypto.PublicKeyFromBytes(p.PublicKey); err != nil {
		return err
	}

	if sig, err = crypto.SignatureFromBytes(p.Signature); err != nil {
		return err
	}

	ppm := new(PbftPrePrepareMessage)
	if err = ppm.FromProto(p.PrePrepareMessage); err != nil {
		return err
	}

	vcm := make([]*PbftViewChangeMessage, len(p.ViewChangeMessages))
	for i, vcProto := range p.ViewChangeMessages {
		vcm[i] = new(PbftViewChangeMessage)
		if err = vcm[i].FromProto(vcProto); err != nil {
			return err
		}
	}

	m.NewView = p.NewView
	m.Sequence = p.Sequence
	m.PrePrepareMessage = ppm
	m.ViewChangeMessages = vcm
	m.PublicKey = key
	m.Signature = sig

	return nil
}

func (m *PbftNewViewMessage) EmptyProto() proto.Message {
	return &pbPbft.PbftNewViewMessage{}
}

func (m *PbftNewViewMessage) Address() types.Address {
	return m.PublicKey.Address()
}
