package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/protomesh/go-app"
	"github.com/protomesh/protomesh/pkg/gateway"
	"github.com/protomesh/protomesh/pkg/pubsub"
	typesv1 "github.com/protomesh/protomesh/proto/api/types/v1"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

var (
	_      gateway.GatewayHandler = &LambdaGatewayHandler[LambdaGatewayHandlerDependency]{}
	nounce                        = 8
)

type LambdaNotification struct {
	Service string `json:"service"`
	Method  string `json:"method"`
}

func LambdaNotificationRedisDeserializer(redisMsg *redis.Message) *pubsub.Message[*LambdaNotification] {

	notification := &LambdaNotification{}

	err := json.Unmarshal([]byte(redisMsg.Payload), notification)
	if err != nil {
		panic(err)
	}

	return &pubsub.Message[*LambdaNotification]{
		Topic:   redisMsg.Channel,
		Payload: notification,
	}

}

type LambdaGatewayHandlerDependency interface {
	GetLambdaClient() *lambda.Client
}

type LambdaGatewayHandler[D LambdaGatewayHandlerDependency] struct {
	*app.Injector[D]

	LambdaStreamSubscriber  pubsub.PubSubSubscriber[*LambdaNotification]
	ServerStreamTopicPrefix app.Config `config:"server.stream.topic.prefix,str" default:"lambda.server.stream:" usage:"Server side Redis channel prefix for pubsub notification"`

	GrpcServerStreamTimeoutHeader app.Config `config:"grpc.server.stream.timeout.header,str" default:"x-server-stream-timeout" usage:"Timeout for waiting for gRPC stream to finish"`

	NounceHeader app.Config `config:"nounce.header,str" default:"x-nounce" usage:"Header to use for nounce"`
	nounceHeader string

	subscriberId uint64

	lambdaCli *lambda.Client
}

func (l *LambdaGatewayHandler[D]) Initialize() {
	l.lambdaCli = l.Dependency().GetLambdaClient()
	l.nounceHeader = l.NounceHeader.StringVal()
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

	if len(incomingMetadata.Get(l.nounceHeader)) == 0 {
		incomingMetadata.Set(l.nounceHeader, RandStringBytesMaskImprSrcSB(nounce))
	}

	serverStream := grpc.ServerTransportStreamFromContext(call.Stream.Context())

	fullPath := serverStream.Method()

	handler := &lambdaGrpcHandler{
		log:                       l.Log().With("source", "gRPC", "fullPath", fullPath),
		fullPath:                  fullPath,
		param:                     param.(*typesv1.AwsHandler_LambdaFunction),
		lambdaCli:                 l.lambdaCli,
		ctx:                       ctx,
		incomingMetadata:          incomingMetadata,
		serverStreamTimeoutHeader: l.GrpcServerStreamTimeoutHeader.StringVal(),
		waitCall:                  make(chan interface{}),
	}

	signalHeader := ""

	if len(call.Policy.LambdaStreamSignalHeaderKey) > 0 {

		signalHeaderMeta := incomingMetadata.Get(call.Policy.LambdaStreamSignalHeaderKey)
		if len(signalHeaderMeta) > 0 {
			signalHeader = signalHeaderMeta[0]
		}

	}

	if l.LambdaStreamSubscriber != nil && len(signalHeader) > 0 {

		topic := fmt.Sprintf("%s%s", l.ServerStreamTopicPrefix.StringVal(), signalHeader)

		subscriber := l.LambdaStreamSubscriber.Subscribe(atomic.AddUint64(&l.subscriberId, 1), topic)

		handler.attachPubSub(l.LambdaStreamSubscriber, subscriber)
	}

	return handler

}

func (l *LambdaGatewayHandler[D]) HandleHttp(ctx context.Context, param proto.Message, call *gateway.HttpCall) gateway.HttpHandler {

	if len(call.Request.Header.Get(l.nounceHeader)) == 0 {
		call.Request.Header.Set(l.nounceHeader, RandStringBytesMaskImprSrcSB(nounce))
	}

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
