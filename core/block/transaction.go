package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/codec"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
	pb "github.com/andantan/p2p-pbft-modular-blockchain-network/proto/block/transaction"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/types"
	"google.golang.org/protobuf/proto"
	"sync"
	"time"
)

type Transaction struct {
	Data      []byte
	From      *crypto.PublicKey
	Signature *crypto.Signature
	Nonce     uint64

	hashLock  sync.RWMutex
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
	tx.hashLock.RLock()
	if !tx.hash.IsZero() {
		defer tx.hashLock.RUnlock()
		return tx.hash, nil
	}
	tx.hashLock.RUnlock()

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
	hash := sha256.Sum256(b)

	tx.hashLock.Lock()
	defer tx.hashLock.Unlock()

	if !tx.hash.IsZero() {
		return tx.hash, nil
	}
	tx.hash = hash

	return tx.hash, nil
}

func (tx *Transaction) dataHash() types.Hash {
	buf := new(bytes.Buffer)

	_ = binary.Write(buf, binary.LittleEndian, tx.Data)
	_ = binary.Write(buf, binary.LittleEndian, tx.Nonce)

	return sha256.Sum256(buf.Bytes())
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

func (tx *Transaction) SetFirstSeen() {
	tx.firstSeen = time.Now().UnixNano()
}

func (tx *Transaction) FirstSeen() int64 {
	return tx.firstSeen
}
