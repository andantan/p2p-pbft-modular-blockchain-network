package block

import (
	"crypto/sha256"
	"errors"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/codec"
	pb "github.com/andantan/p2p-pbft-modular-blockchain-network/proto/core/block"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/types"
	"google.golang.org/protobuf/proto"
	"time"
)

const (
	Version uint32 = 1
)

type Header struct {
	Version       uint32
	MerkleRoot    types.Hash
	PrevBlockHash types.Hash
	Timestamp     int64
	Height        uint64
	Weight        uint64

	StateRoot types.Hash
	Nonce     uint64
}

func NewHeader(m types.Hash, ph types.Hash, h uint64, w uint64, s types.Hash, n uint64) *Header {
	return &Header{
		Version:       Version,
		MerkleRoot:    m,
		PrevBlockHash: ph,
		Timestamp:     time.Now().UnixNano(),
		Height:        h,
		Weight:        w,
		StateRoot:     s,
		Nonce:         n,
	}
}

func (h *Header) Hash() (types.Hash, error) {
	b, err := codec.EncodeProto(h)
	if err != nil {
		return types.Hash{}, err
	}

	hash := sha256.Sum256(b)

	return hash, nil
}

func (h *Header) ToProto() (proto.Message, error) {
	return &pb.Header{
		Version:       h.Version,
		MerkleRoot:    h.MerkleRoot.Bytes(),
		PrevBlockHash: h.PrevBlockHash.Bytes(),
		Timestamp:     h.Timestamp,
		Height:        h.Height,
		Weight:        h.Weight,
		StateRoot:     h.StateRoot.Bytes(),
		Nonce:         h.Nonce,
	}, nil
}

func (h *Header) FromProto(msg proto.Message) error {
	p, ok := msg.(*pb.Header)
	if !ok {
		return errors.New("invalid proto message type for Header")
	}

	var (
		err           error
		merkleRoot    types.Hash
		prevBlockHash types.Hash
		stateRoot     types.Hash
	)

	if merkleRoot, err = types.HashFromBytes(p.MerkleRoot); err != nil {
		return err
	}

	if prevBlockHash, err = types.HashFromBytes(p.PrevBlockHash); err != nil {
		return err
	}

	if stateRoot, err = types.HashFromBytes(p.StateRoot); err != nil {
		return err
	}

	h.Version = p.Version
	h.MerkleRoot = merkleRoot
	h.PrevBlockHash = prevBlockHash
	h.Timestamp = p.Timestamp
	h.Height = p.Height
	h.Weight = p.Weight
	h.StateRoot = stateRoot
	h.Nonce = p.Nonce

	return nil
}

func (h *Header) EmptyProto() proto.Message {
	return &pb.Header{}
}
