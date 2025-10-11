package tcp

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/andantan/modular-blockchain/crypto"
	pb "github.com/andantan/modular-blockchain/proto/network/protocol/tcp"
	"github.com/andantan/modular-blockchain/types"
	"google.golang.org/protobuf/proto"
)

type TcpHandshakeMessage struct {
	PublicKey *crypto.PublicKey
	NetAddr   string
	Signature *crypto.Signature
}

func NewTcpHandshakeMessage(pubKey *crypto.PublicKey, netAddr string) *TcpHandshakeMessage {
	return &TcpHandshakeMessage{
		PublicKey: pubKey,
		NetAddr:   netAddr,
	}
}

func (h *TcpHandshakeMessage) Hash() (types.Hash, error) {
	if h.PublicKey == nil {
		return types.Hash{}, fmt.Errorf("cannot hash identity with nil public key")
	}

	buf := new(bytes.Buffer)
	buf.Write(h.PublicKey.Key)
	buf.Write([]byte(h.NetAddr))
	hash := sha256.Sum256(buf.Bytes())

	return hash, nil
}

func (h *TcpHandshakeMessage) Sign(privKey *crypto.PrivateKey) error {
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

func (h *TcpHandshakeMessage) Verify() error {
	if h.Signature == nil {
		return fmt.Errorf("TcpHandshakeMessage has no signature to verify")
	}
	if h.PublicKey == nil {
		return fmt.Errorf("TcpHandshakeMessage has no public key to verify with")
	}

	hash, err := h.Hash()

	if err != nil {
		return err
	}

	if !h.Signature.Verify(h.PublicKey, hash.Bytes()) {
		return fmt.Errorf("invalid TcpHandshakeMessage signature")
	}

	return nil
}

func (h *TcpHandshakeMessage) ToProto() (proto.Message, error) {
	return &pb.TcpHandshakeMessage{
		PublicKey: h.PublicKey.Bytes(),
		NetAddr:   h.NetAddr,
		Signature: h.Signature.Bytes(),
	}, nil
}

func (h *TcpHandshakeMessage) FromProto(msg proto.Message) error {
	p, ok := msg.(*pb.TcpHandshakeMessage)
	if !ok {
		return errors.New("invalid proto message type for TcpHandshakeMessage")
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

func (h *TcpHandshakeMessage) EmptyProto() proto.Message {
	return &pb.TcpHandshakeMessage{}
}

type TcpRawMessage struct {
	protocol string
	from     types.Address
	payload  []byte
}

func NewTcpRawMessage(from types.Address, payload []byte) *TcpRawMessage {
	return &TcpRawMessage{
		protocol: "tcp",
		from:     from,
		payload:  payload,
	}
}

func (m *TcpRawMessage) Protocol() string {
	return m.protocol
}

func (m *TcpRawMessage) From() types.Address {
	return m.from
}

func (m *TcpRawMessage) Payload() []byte {
	return m.payload
}
