package block

import (
	"fmt"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
	pb "github.com/andantan/p2p-pbft-modular-blockchain-network/proto/core/block"
	"google.golang.org/protobuf/proto"
)

type CommitVote struct {
	PublicKey *crypto.PublicKey
	Signature *crypto.Signature
}

func NewCommitVote(key *crypto.PublicKey, sig *crypto.Signature) *CommitVote {
	return &CommitVote{
		PublicKey: key,
		Signature: sig,
	}
}

func (cv *CommitVote) ToProto() (proto.Message, error) {
	if cv.PublicKey == nil || cv.Signature == nil {
		return nil, fmt.Errorf("commit vote public key or signature is nil")
	}

	return &pb.CommitVote{
		PublicKey: cv.PublicKey.Bytes(),
		Signature: cv.Signature.Bytes(),
	}, nil
}

func (cv *CommitVote) FromProto(msg proto.Message) error {
	p, ok := msg.(*pb.CommitVote)
	if !ok {
		return fmt.Errorf("invalid proto message type for CommitVote")
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

	cv.PublicKey = key
	cv.Signature = sig

	return nil
}

func (cv *CommitVote) EmptyProto() proto.Message {
	return &pb.CommitVote{}
}

type Tail struct {
	CommitVotes []*CommitVote
}

func NewTail(cvs []*CommitVote) *Tail {
	return &Tail{
		CommitVotes: cvs,
	}
}

func (t *Tail) ToProto() (proto.Message, error) {
	commitsProto := make([]*pb.CommitVote, len(t.CommitVotes))
	for i, commit := range t.CommitVotes {
		commitProto, err := commit.ToProto()
		if err != nil {
			return nil, err
		}
		commitsProto[i] = commitProto.(*pb.CommitVote)
	}

	return &pb.Tail{
		CommitVotes: commitsProto,
	}, nil
}

func (t *Tail) FromProto(msg proto.Message) error {
	p, ok := msg.(*pb.Tail)
	if !ok {
		return fmt.Errorf("invalid proto message type for Tail")
	}

	t.CommitVotes = make([]*CommitVote, len(p.CommitVotes))
	for i, v := range p.CommitVotes {
		t.CommitVotes[i] = new(CommitVote)
		if err := t.CommitVotes[i].FromProto(v); err != nil {
			return err
		}
	}

	return nil
}

func (t *Tail) EmptyProto() proto.Message {
	return &pb.Tail{}
}
