package codec

import "google.golang.org/protobuf/proto"

type ProtoCodec interface {
	ToProto() (proto.Message, error)
	FromProto(proto.Message) error
	EmptyProto() proto.Message
}

func EncodeProto(c ProtoCodec) ([]byte, error) {
	pb, err := c.ToProto()
	if err != nil {
		return nil, err
	}
	return proto.Marshal(pb)
}

func DecodeProto(b []byte, c ProtoCodec) error {
	dest := c.EmptyProto()

	if err := proto.Unmarshal(b, dest); err != nil {
		return err
	}

	return c.FromProto(dest)
}
