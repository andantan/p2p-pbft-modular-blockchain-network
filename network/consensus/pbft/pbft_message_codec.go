package pbft

import (
	"errors"
	"fmt"
	"github.com/andantan/modular-blockchain/codec"
	"github.com/andantan/modular-blockchain/network/consensus"
	"github.com/andantan/modular-blockchain/network/message"
)

const (
	MessagePbftConsensusSubTypePrePrepare message.MessageConsensusSubType = iota
	MessagePbftConsensusSubTypePrepare
	MessagePbftConsensusSubTypeCommit
	MessagePbftConsensusSubTypeViewChange
	MessagePbftConsensusSubTypeNewView
)

type PbftConsensusMessageCodec struct{}

func NewPbftConsensusMessageCodec() *PbftConsensusMessageCodec {
	return &PbftConsensusMessageCodec{}
}

func (c *PbftConsensusMessageCodec) Encode(msg consensus.ConsensusMessage) ([]byte, error) {
	payload := make([]byte, 0)
	payload = append(payload, byte(message.MessageConsensusType))

	var subType message.MessageConsensusSubType
	switch msg.(type) {
	case *PbftPrePrepareMessage:
		subType = MessagePbftConsensusSubTypePrePrepare
	case *PbftPrepareMessage:
		subType = MessagePbftConsensusSubTypePrepare
	case *PbftCommitMessage:
		subType = MessagePbftConsensusSubTypeCommit
	case *PbftViewChangeMessage:
		subType = MessagePbftConsensusSubTypeViewChange
	case *PbftNewViewMessage:
		subType = MessagePbftConsensusSubTypeNewView
	}

	payload = append(payload, byte(subType))

	data, err := codec.EncodeProto(msg)
	if err != nil {
		return nil, err
	}

	payload = append(payload, data...)

	return payload, nil
}

func (c *PbftConsensusMessageCodec) Decode(payload []byte) (consensus.ConsensusMessage, error) {
	if len(payload) < 2 {
		return nil, errors.New("payload too short")
	}

	msgType := message.MessageType(payload[0])
	if msgType != message.MessageConsensusType {
		return nil, fmt.Errorf("invalid message type: expected Consensus, got %d", msgType)
	}

	subType := message.MessageConsensusSubType(payload[1])
	data := payload[2:]

	var concreteMsg consensus.ConsensusMessage
	switch subType {
	case MessagePbftConsensusSubTypePrePrepare:
		concreteMsg = new(PbftPrePrepareMessage)
	case MessagePbftConsensusSubTypePrepare:
		concreteMsg = new(PbftPrepareMessage)
	case MessagePbftConsensusSubTypeCommit:
		concreteMsg = new(PbftCommitMessage)
	case MessagePbftConsensusSubTypeViewChange:
		concreteMsg = new(PbftViewChangeMessage)
	case MessagePbftConsensusSubTypeNewView:
		concreteMsg = new(PbftNewViewMessage)
	default:
		return nil, fmt.Errorf("unknown consensus message subtype: %d", subType)
	}

	if err := codec.DecodeProto(data, concreteMsg); err != nil {
		return nil, err
	}

	return concreteMsg, nil
}
