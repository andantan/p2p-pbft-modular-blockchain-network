package block

import (
	"crypto/sha256"
	"errors"
	"fmt"
	pb "github.com/andantan/p2p-pbft-modular-blockchain-network/proto/core/block"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/types"
	"google.golang.org/protobuf/proto"
	"sort"
)

type Body struct {
	Transactions []*Transaction
}

func NewBody(txx []*Transaction) *Body {
	if txx == nil {
		txx = make([]*Transaction, 0)
	}

	return &Body{
		Transactions: txx,
	}
}

func (b *Body) Weight() uint64 {
	return uint64(len(b.Transactions))
}

func (b *Body) CalculateMerkleRoot() (types.Hash, error) {
	if b.Transactions == nil {
		return types.Hash{}, errors.New("transactions is nil")
	}

	if len(b.Transactions) == 0 {
		return types.ZeroHash, nil
	}

	sort.Slice(b.Transactions, func(i, j int) bool {
		txHashA, _ := b.Transactions[i].Hash()
		txHashB, _ := b.Transactions[j].Hash()

		return txHashA.Lte(txHashB)
	})

	var hashes []types.Hash
	for _, tx := range b.Transactions {
		hash, _ := tx.Hash()
		hashes = append(hashes, hash)
	}

	for len(hashes) > 1 {
		if len(hashes)%2 != 0 {
			hashes = append(hashes, hashes[len(hashes)-1])
		}

		var nextLevelHashes []types.Hash

		for i := 0; i < len(hashes); i += 2 {
			left := hashes[i]
			right := hashes[i+1]

			combinedHashData := append(left.Bytes(), right.Bytes()...)
			parentHash := types.Hash(sha256.Sum256(combinedHashData))

			nextLevelHashes = append(nextLevelHashes, parentHash)
		}

		hashes = nextLevelHashes
	}

	return hashes[0], nil
}

func (b *Body) ToProto() (proto.Message, error) {
	txxProto, err := TransactionsToProto(b.Transactions)

	if err != nil {
		return nil, err
	}

	return &pb.Body{
		Transactions: txxProto,
	}, nil
}

func (b *Body) FromProto(msg proto.Message) error {
	p, ok := msg.(*pb.Body)
	if !ok {
		return fmt.Errorf("invalid proto message type for Body")
	}

	var (
		err error
		txx []*Transaction
	)

	if txx, err = TransactionsFromProto(p.Transactions); err != nil {
		return err
	}

	b.Transactions = txx
	return nil
}

func (b *Body) EmptyProto() proto.Message {
	return &pb.Body{}
}
