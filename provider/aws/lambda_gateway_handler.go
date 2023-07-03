package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/protomesh/go-app"
	"github.com/protomesh/protomesh/pkg/gateway"
	typesv1 "github.com/protomesh/protomesh/proto/api/types/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

var (
	_ gateway.GatewayHandler = &LambdaGatewayHandler[LambdaGatewayHandlerDependency]{}
)

type LambdaGatewayHandlerDependency interface {
	GetLambdaClient() *lambda.Client
}

type LambdaGatewayHandler[D LambdaGatewayHandlerDependency] struct {
	*app.Injector[D]

	lambdaCli *lambda.Client
}

func (l *LambdaGatewayHandler[D]) Initialize() {
	l.lambdaCli = l.Dependency().GetLambdaClient()
}

func (l *LambdaGatewayHandler[D]) GetHandlerType() gateway.HandlerType {
	return gateway.HandlerTypeAwsLambda
}

func (l *LambdaGatewayHandler[D]) ProcessPolicies(ctx context.Context, updated []*typesv1.GatewayPolicy, dropped []*typesv1.GatewayPolicy) error {
	// For lambda, there's no need to eager process policies
	return nil

}

func (l *LambdaGatewayHandler[D]) HandleGrpc(ctx context.Context, param proto.Message, call *gateway.GrpcCall) gateway.GrpcHandler {

	incomingMetadata, _ := metadata.FromIncomingContext(ctx)

	serverStream := grpc.ServerTransportStreamFromContext(call.Stream.Context())

	fullPath := serverStream.Method()

	handler := &lambdaGrpcHandler{
		log:              l.Log().With("source", "gRPC", "fullPath", fullPath),
		fullPath:         fullPath,
		param:            param.(*typesv1.AwsHandler_LambdaFunction),
		lambdaCli:        l.lambdaCli,
		ctx:              ctx,
		incomingMetadata: incomingMetadata,
		waitCall:         make(chan interface{}),
	}

	return handler

}

func (l *LambdaGatewayHandler[D]) HandleHttp(ctx context.Context, param proto.Message, call *gateway.HttpCall) gateway.HttpHandler {

	handler := &lambdaHttpHandler{
		log:       l.Log().With("source", "HTTP", "path", call.Request.URL.Path),
		param:     param.(*typesv1.AwsHandler_LambdaFunction),
		lambdaCli: l.lambdaCli,
		ctx:       ctx,
		request:   call.Request,
		response:  call.Response,
	}

	return handler

}
