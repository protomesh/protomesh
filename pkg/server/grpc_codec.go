package server

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

func Codec() grpc.Codec {
	return CodecWithParent(&protoCodec{})
}

func CodecWithParent(fallback grpc.Codec) grpc.Codec {
	return &rawCodec{fallback}
}

type rawCodec struct {
	parentCodec grpc.Codec
}

type frame struct {
	payload []byte
}

func newFrame(payload []byte) *frame {
	return &frame{payload: payload}
}

func (c *rawCodec) Marshal(v interface{}) ([]byte, error) {
	out, ok := v.(*frame)
	if !ok {
		return c.parentCodec.Marshal(v)
	}
	return out.payload, nil
}

func (c *rawCodec) Unmarshal(data []byte, v interface{}) error {
	dst, ok := v.(*frame)
	if !ok {
		return c.parentCodec.Unmarshal(data, v)
	}
	dst.payload = data
	return nil
}

func (c *rawCodec) String() string {
	return fmt.Sprintf("proxy>%s", c.parentCodec.String())
}

type protoCodec struct{}

func (protoCodec) Marshal(v interface{}) ([]byte, error) {
	vv, ok := v.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("failed to marshal, message is %T, want proto.Message", v)
	}
	return proto.Marshal(vv)
}

func (protoCodec) Unmarshal(data []byte, v interface{}) error {
	vv, ok := v.(proto.Message)
	if !ok {
		return fmt.Errorf("failed to unmarshal, message is %T, want proto.Message", v)
	}
	return proto.Unmarshal(data, vv)
}

func (protoCodec) String() string {
	return "proto"
}
