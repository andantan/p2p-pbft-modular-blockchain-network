package message

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/core/block"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
	pbBlock "github.com/andantan/p2p-pbft-modular-blockchain-network/proto/core/block"
	pbPbft "github.com/andantan/p2p-pbft-modular-blockchain-network/proto/network/message"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/types"
	"google.golang.org/protobuf/proto"
)

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
