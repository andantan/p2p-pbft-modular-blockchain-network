package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/codec"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
	pb "github.com/andantan/p2p-pbft-modular-blockchain-network/proto/core/block"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/types"
	"google.golang.org/protobuf/proto"
	"time"
)

type Transaction struct {
	Data      []byte
	From      *crypto.PublicKey
	Signature *crypto.Signature
	Nonce     uint64

	hash      types.Hash
	firstSeen int64
}

func NewTransaction(data []byte, nonce uint64) *Transaction {
	return &Transaction{
		Data:  data,
		Nonce: nonce,
	}
}

func (tx *Transaction) Hash() (types.Hash, error) {
	if !tx.hash.IsZero() {
		return tx.hash, nil
	}

	if tx.Signature == nil || tx.From == nil {
		return types.Hash{}, fmt.Errorf("cannot hash an unsigned transaction")
	}

	if tx.Signature.IsNil() {
		return types.Hash{}, fmt.Errorf("cannot hash an unvalid R, S signature in transaction")
	}

	b, err := codec.EncodeProto(tx)
	if err != nil {
		return types.Hash{}, err
	}

	tx.hash = sha256.Sum256(b)

	return tx.hash, nil
}

func (tx *Transaction) Sign(privKey *crypto.PrivateKey) error {
	hash := tx.dataHash()

	sig, err := privKey.Sign(hash.Bytes())
	if err != nil {
		return err
	}

	tx.From = privKey.PublicKey()
	tx.Signature = sig

	return nil
}

func (tx *Transaction) Verify() error {
	if tx.Signature == nil || tx.From == nil {
		return fmt.Errorf("transaction has no signature or sender")
	}

	if tx.Signature.IsNil() {
		return fmt.Errorf("transaction has no signature R or S")
	}

	hash := tx.dataHash()

	if !tx.Signature.Verify(tx.From, hash.Bytes()) {
		return fmt.Errorf("invalid transaction signature")
	}

	return nil
}

func (tx *Transaction) ToProto() (proto.Message, error) {
	return &pb.Transaction{
		Data:      tx.Data,
		From:      tx.From.Bytes(),
		Signature: tx.Signature.Bytes(),
		Nonce:     tx.Nonce,
	}, nil
}

func (tx *Transaction) FromProto(msg proto.Message) error {
	p, ok := msg.(*pb.Transaction)
	if !ok {
		return fmt.Errorf("invalid proto message type for Transaction")
	}

	var (
		err error
		key *crypto.PublicKey
		sig *crypto.Signature
	)

	if key, err = crypto.PublicKeyFromBytes(p.From); err != nil {
		return err
	}

	if sig, err = crypto.SignatureFromBytes(p.Signature); err != nil {
		return err
	}

	tx.Data = p.Data
	tx.From = key
	tx.Signature = sig
	tx.Nonce = p.Nonce

	return nil
}

func (tx *Transaction) EmptyProto() proto.Message {
	return &pb.Transaction{}
}

func (tx *Transaction) FirstSeen() int64 {
	if tx.firstSeen == 0 {
		tx.firstSeen = time.Now().UnixNano()
	}

	return tx.firstSeen
}

func (tx *Transaction) dataHash() types.Hash {
	buf := new(bytes.Buffer)

	_ = binary.Write(buf, binary.LittleEndian, tx.Data)
	_ = binary.Write(buf, binary.LittleEndian, tx.Nonce)

	return sha256.Sum256(buf.Bytes())
}

func TransactionsToProto(txx []*Transaction) ([]*pb.Transaction, error) {
	protoTxx := make([]*pb.Transaction, len(txx))
	for i, tx := range txx {
		protoTx, err := tx.ToProto()
		if err != nil {
			return nil, err
		}
		protoTxx[i] = protoTx.(*pb.Transaction)
	}

	return protoTxx, nil
}

func TransactionsFromProto(protoTxx []*pb.Transaction) ([]*Transaction, error) {
	txx := make([]*Transaction, len(protoTxx))
	for i, protoTx := range protoTxx {
		tx := new(Transaction)
		if err := tx.FromProto(protoTx); err != nil {
			return nil, err
		}
		txx[i] = tx
	}

	return txx, nil
}
