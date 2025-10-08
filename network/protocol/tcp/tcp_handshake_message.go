package tcp

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
	pb "github.com/andantan/p2p-pbft-modular-blockchain-network/proto/network/protocol/tcp"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/types"
	"google.golang.org/protobuf/proto"
)

type TCPHandshakeMessage struct {
	PublicKey *crypto.PublicKey
	NetAddr   string
	Signature *crypto.Signature
}

func NewTCPHandshakeMessage(pubKey *crypto.PublicKey, netAddr string) *TCPHandshakeMessage {
	return &TCPHandshakeMessage{
		PublicKey: pubKey,
		NetAddr:   netAddr,
	}
}

func (h *TCPHandshakeMessage) Hash() (types.Hash, error) {
	if h.PublicKey == nil {
		return types.Hash{}, fmt.Errorf("cannot hash identity with nil public key")
	}

	buf := new(bytes.Buffer)
	buf.Write(h.PublicKey.Key)
	buf.Write([]byte(h.NetAddr))
	hash := sha256.Sum256(buf.Bytes())

	return hash, nil
}

func (h *TCPHandshakeMessage) Sign(privKey *crypto.PrivateKey) error {
	hash, err := h.Hash()

	if err != nil {
		return err
	}

	sig, err := privKey.Sign(hash.Bytes())

	if err != nil {
		return err
	}

	h.Signature = sig
	return nil
}

func (h *TCPHandshakeMessage) Verify() error {
	if h.Signature == nil {
		return fmt.Errorf("TCPHandshakeMessage has no signature to verify")
	}
	if h.PublicKey == nil {
		return fmt.Errorf("TCPHandshakeMessage has no public key to verify with")
	}

	hash, err := h.Hash()

	if err != nil {
		return err
	}

	if !h.Signature.Verify(h.PublicKey, hash.Bytes()) {
		return fmt.Errorf("invalid TCPHandshakeMessage signature")
	}

	return nil
}

func (h *TCPHandshakeMessage) ToProto() (proto.Message, error) {
	return &pb.TcpHandshakeMessage{
		PublicKey: h.PublicKey.Bytes(),
		NetAddr:   h.NetAddr,
		Signature: h.Signature.Bytes(),
	}, nil
}

func (h *TCPHandshakeMessage) FromProto(msg proto.Message) error {
	p, ok := msg.(*pb.TcpHandshakeMessage)
	if !ok {
		return errors.New("invalid proto message type for TCPHandshakeMessage")
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

	h.PublicKey = key
	h.NetAddr = p.NetAddr
	h.Signature = sig

	return nil
}

func (h *TCPHandshakeMessage) EmptyProto() proto.Message {
	return &pb.TcpHandshakeMessage{}
}
