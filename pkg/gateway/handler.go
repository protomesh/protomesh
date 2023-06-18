package gateway

import (
	"context"
	"net/http"

	typesv1 "github.com/protomesh/protomesh/proto/api/types/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

type GrpcCall struct {
	Policy   *typesv1.GatewayPolicy
	Stream   grpc.ServerStream
	Handlers []GrpcHandler
}

type GrpcHandler interface {
	Call([]byte) error
	Result() ([]byte, error)
	GetOutgoingMetadata() metadata.MD
}

type HttpCall struct {
	Policy   *typesv1.GatewayPolicy
	Response http.ResponseWriter
	Request  *http.Request
	Handlers []HttpHandler
}

type HttpHandler interface {
	Call() error
}

type GatewayHandler interface {
	GetHandlerType() HandlerType
	ProcessPolicies(context.Context, []*typesv1.GatewayPolicy, []*typesv1.GatewayPolicy) error

	HandleGrpc(context.Context, proto.Message, *GrpcCall) GrpcHandler
	HandleHttp(context.Context, proto.Message, *HttpCall) HttpHandler
}

type HandlerType string

const (
	HandlerTypeUndefined HandlerType = ""
	HandlerTypeAwsLambda HandlerType = "aws_lambda"
)

type handlerMatcher map[HandlerType]GatewayHandler

func newHandlerMatcher() handlerMatcher {
	return make(handlerMatcher)
}

func (hm handlerMatcher) getHandler(h HandlerType) GatewayHandler {
	handler, ok := hm[h]
	if ok {
		return handler
	}
	return nil
}

func (hm handlerMatcher) fromHandlersSlice(handlers []GatewayHandler) {
	for _, handler := range handlers {
		hm[handler.GetHandlerType()] = handler
	}
}
