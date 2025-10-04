package network

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
	pb "github.com/andantan/p2p-pbft-modular-blockchain-network/proto/handshake"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/types"
	"google.golang.org/protobuf/proto"
)

type HandshakeMessage struct {
	PublicKey *crypto.PublicKey
	NetAddr   string
	Signature *crypto.Signature
}

func NewHandshakeMessage(pubKey *crypto.PublicKey, netAddr string) *HandshakeMessage {
	return &HandshakeMessage{
		PublicKey: pubKey,
		NetAddr:   netAddr,
	}
}

func (h *HandshakeMessage) Hash() (types.Hash, error) {
	if h.PublicKey == nil {
		return types.Hash{}, fmt.Errorf("cannot hash identity with nil public key")
	}

	buf := new(bytes.Buffer)
	buf.Write(h.PublicKey.Key)
	buf.Write([]byte(h.NetAddr))
	hash := sha256.Sum256(buf.Bytes())

	return hash, nil
}

func (h *HandshakeMessage) Sign(privKey *crypto.PrivateKey) error {
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

func (h *HandshakeMessage) Verify() error {
	if h.Signature == nil {
		return fmt.Errorf("peerIdentity has no signature to verify")
	}
	if h.PublicKey == nil {
		return fmt.Errorf("peerIdentity has no public key to verify with")
	}

	hash, err := h.Hash()

	if err != nil {
		return err
	}

	if !h.Signature.Verify(h.PublicKey, hash.Bytes()) {
		return fmt.Errorf("invalid peerIdentity signature")
	}

	return nil
}

func (h *HandshakeMessage) ToProto() (proto.Message, error) {
	return &pb.Handshake{
		PublicKey: h.PublicKey.Bytes(),
		NetAddr:   h.NetAddr,
		Signature: h.Signature.Bytes(),
	}, nil
}

func (h *HandshakeMessage) FromProto(msg proto.Message) error {
	p, ok := msg.(*pb.Handshake)
	if !ok {
		return fmt.Errorf("invalid proto message type for PeerIdentity")
	}

	pubKey, err := crypto.PublicKeyFromBytes(p.PublicKey)
	if err != nil {
		return err
	}
	sig, err := crypto.SignatureFromBytes(p.Signature)
	if err != nil {
		return err
	}

	h.PublicKey = pubKey
	h.NetAddr = p.NetAddr
	h.Signature = sig

	return nil
}

func (h *HandshakeMessage) EmptyProto() proto.Message {
	return &pb.Handshake{}
}
