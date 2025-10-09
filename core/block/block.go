package block

import (
	"errors"
	"fmt"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
	pb "github.com/andantan/p2p-pbft-modular-blockchain-network/proto/core/block"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/types"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/util"
	"google.golang.org/protobuf/proto"
)

type Block struct {
	blockHash types.Hash

	Header *Header
	Body   *Body
	Tail   *Tail

	Proposer  *crypto.PublicKey
	Signature *crypto.Signature
}

func NewBlock(header *Header, body *Body) (*Block, error) {
	b := &Block{
		Header: header,
		Body:   body,
	}

	h, err := b.Hash()

	if err != nil {
		return nil, err
	}

	b.blockHash = h

	return b, nil
}

func NewBlockFromPrevHeader(prevHeader *Header, newBody *Body) (*Block, error) {
	m, err := newBody.Hash()

	if err != nil {
		return nil, err
	}

	ph, err := prevHeader.Hash()

	if err != nil {
		return nil, err
	}

	ht := prevHeader.Height + 1
	wt := newBody.GetWeight()
	s := types.ZeroHash
	n := util.RandomUint64()

	h := NewHeader(m, ph, ht, wt, s, n)

	return NewBlock(h, newBody)
}

func NewGenesisBlock() *Block {
	h := &Header{
		Version:       0,
		MerkleRoot:    types.ZeroHash,
		PrevBlockHash: types.FFHash,
		Timestamp:     0,
		Height:        0,
		Weight:        0,
		StateRoot:     types.ZeroHash,
		Nonce:         0,
	}

	bd := &Body{
		Transactions: []*Transaction{},
	}

	b, err := NewBlock(h, bd)

	p, err := crypto.GeneratePrivateKey()

	if err != nil {
		panic(err)
	}

	if err = b.Sign(p); err != nil {
		panic(err)
	}

	if err = b.Verify(); err != nil {
		panic(err)
	}

	return b
}

func (b *Block) Hash() (types.Hash, error) {
	if !b.blockHash.IsZero() {
		return b.blockHash, nil
	}

	h, err := b.Header.Hash()

	if err != nil {
		return types.Hash{}, err
	}
	b.blockHash = h

	return h, nil
}

func (b *Block) Sign(privKey *crypto.PrivateKey) error {
	if b.blockHash.IsZero() {
		return errors.New("block hash is zero")
	}

	sig, err := privKey.Sign(b.blockHash.Bytes())

	if err != nil {
		return err
	}

	b.Proposer = privKey.PublicKey()
	b.Signature = sig

	return nil
}

func (b *Block) Verify() error {
	if b.Signature == nil || b.Proposer == nil {
		return fmt.Errorf("block has no signature or sender")
	}

	if b.Signature.IsNil() {
		return fmt.Errorf("block has no signature R or S")
	}

	if !b.Signature.Verify(b.Proposer, b.blockHash.Bytes()) {
		return fmt.Errorf("block signature is invalid")
	}

	return nil
}

func (b *Block) ToProto() (proto.Message, error) {
	var (
		err         error
		headerProto proto.Message
		bodyProto   proto.Message
		tailproto   proto.Message
	)

	if headerProto, err = b.Header.ToProto(); err != nil {
		return nil, err
	}

	if bodyProto, err = b.Body.ToProto(); err != nil {
		return nil, err
	}

	var tp *pb.Tail
	if b.Tail != nil {
		if tailproto, err = b.Tail.ToProto(); err != nil {
			return nil, err
		}
		tp = tailproto.(*pb.Tail)
	}

	return &pb.Block{
		BlockHash: b.blockHash.Bytes(),
		Header:    headerProto.(*pb.Header),
		Body:      bodyProto.(*pb.Body),
		Tail:      tp,
		Proposer:  b.Proposer.Bytes(),
		Signature: b.Signature.Bytes(),
	}, nil
}

func (b *Block) FromProto(msg proto.Message) error {
	p, ok := msg.(*pb.Block)
	if !ok {
		return fmt.Errorf("invalid proto message type for Block")
	}

	b.Header = &Header{}
	if err := b.Header.FromProto(p.Header); err != nil {
		return err
	}

	b.Body = &Body{}
	if err := b.Body.FromProto(p.Body); err != nil {
		return err
	}

	if p.Tail != nil {
		b.Tail = &Tail{}
		if err := b.Tail.FromProto(p.Tail); err != nil {
			return err
		}
	}

	var (
		err       error
		hash      types.Hash
		proposer  *crypto.PublicKey
		signature *crypto.Signature
	)

	if hash, err = b.Header.Hash(); err != nil {
		return err
	}

	if proposer, err = crypto.PublicKeyFromBytes(p.Proposer); err != nil {
		return err
	}

	if signature, err = crypto.SignatureFromBytes(p.Signature); err != nil {
		return err
	}

	b.blockHash = hash
	b.Proposer = proposer
	b.Signature = signature

	return nil
}

func (b *Block) EmptyProto() proto.Message {
	return &pb.Block{}
}

func (b *Block) Seal(votes []*CommitVote, set []types.Address) error {
	if b.Tail != nil {
		return fmt.Errorf("block tail is already set")
	}

	quorum := (2 * len(set) / 3) + 1
	if len(votes) < quorum {
		return fmt.Errorf("not enough commit votes in tail")
	}

	for _, v := range votes {
		found := false
		for _, validatorAddr := range set {
			if validatorAddr.Equal(v.PublicKey.Address()) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("tail contains vote from non-validator: %s", v.PublicKey.Address())
		}

		if !v.Signature.Verify(v.PublicKey, v.Digest.Bytes()) {
			return fmt.Errorf("invalid signature in CommitVote")
		}
	}

	b.Tail = NewTail(votes)

	return nil
}
