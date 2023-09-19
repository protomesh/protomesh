package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/iancoleman/strcase"
	"github.com/protomesh/go-app"
	"github.com/protomesh/protomesh/pkg/client"
	"github.com/protomesh/protomesh/pkg/gateway"
	"github.com/protomesh/protomesh/pkg/pubsub"
	servicesv1 "github.com/protomesh/protomesh/proto/api/services/v1"
	awsprovider "github.com/protomesh/protomesh/provider/aws"
	"github.com/protomesh/protomesh/provider/redis"
	"google.golang.org/grpc"
)

var (
	_ GatewayInjector = &GatewayInstance[*root]{}
)

type GatewayDeps interface {
	GetAwsConfig() aws.Config
	GetGrpcServer() *grpc.Server
}

type GatewayInjector interface {
	GetLambdaClient() *lambda.Client
	GetResourceStoreClient() servicesv1.ResourceStoreClient
}

type GatewayInstance[D GatewayDeps] struct {
	*app.Injector[D]

	ResourceStore       *client.GrpcClient[GatewayInjector] `config:"resource.store"`
	resourceStoreClient servicesv1.ResourceStoreClient

	ctx    context.Context
	cancel context.CancelFunc

	Handlers app.Config `config:"handlers,str" usage:"Enabled gateway handlers, separated by comma"`

	Gateway      *gateway.Gateway[GatewayInjector] `config:"service"`
	gatewayErrCh <-chan error

	AwsLambdaHandler *awsprovider.LambdaGatewayHandler[GatewayInjector] `config:"aws.lambda"`

	RedisClient                *redis.RedisClient[GatewayInjector] `config:"redis"`
	AwsLambdaPubSub            pubsub.PubSub[*awsprovider.LambdaNotification]
	awsLambdaRedisPubSubDriver redis.RedisPubSubDriver
}

func NewGatewayInstance[D GatewayDeps]() *GatewayInstance[D] {
	return &GatewayInstance[D]{
		ResourceStore:    &client.GrpcClient[GatewayInjector]{},
		Gateway:          &gateway.Gateway[GatewayInjector]{},
		AwsLambdaHandler: &awsprovider.LambdaGatewayHandler[GatewayInjector]{},
		RedisClient:      &redis.RedisClient[GatewayInjector]{},
	}
}

func (p *GatewayInstance[D]) GetLambdaClient() *lambda.Client {
	return lambda.NewFromConfig(p.Dependency().GetAwsConfig())
}

func (p *GatewayInstance[D]) GetResourceStoreClient() servicesv1.ResourceStoreClient {
	return p.resourceStoreClient
}

func (p *GatewayInstance[D]) Initialize() {

	log := p.Log()

	p.RedisClient.Initialize()

	handlerTypes := strings.Split(p.Handlers.StringVal(), ",")

	handlers := []gateway.GatewayHandler{}

	for _, handlerType := range handlerTypes {

		switch gateway.HandlerType(strcase.ToSnake(handlerType)) {

		case gateway.HandlerTypeAwsLambda:

			p.AwsLambdaHandler.Initialize()

			streamPrefix := p.AwsLambdaHandler.ServerStreamTopicPrefix.StringVal()

			if p.RedisClient.Client != nil && len(streamPrefix) > 0 {

				p.AwsLambdaPubSub = pubsub.NewPubSub[*awsprovider.LambdaNotification](false, 0, 0)

				p.awsLambdaRedisPubSubDriver = redis.NewRedisPubSubDriver[*awsprovider.LambdaNotification](
					p.RedisClient.Client,
					p.AwsLambdaPubSub,
					awsprovider.LambdaNotificationRedisDeserializer,
				)

				p.AwsLambdaHandler.LambdaStreamSubscriber = p.AwsLambdaPubSub

				log.Info("AWS Lambda stream subscriber enabled")

			}

			log.Info("Initialized AWS Lambda handler")

			handlers = append(handlers, p.AwsLambdaHandler)

		default:
			log.Panic("Unknown gateway handler", "handlerType", handlerType)

		}

	}

	p.Gateway.Initialize(handlers...)

}

func (p *GatewayInstance[D]) Start() {

	p.ctx, p.cancel = context.WithCancel(context.TODO())

	p.ResourceStore.Start()

	p.resourceStoreClient = servicesv1.NewResourceStoreClient(p.ResourceStore.ClientConn)

	p.gatewayErrCh = p.Gateway.Sync(p.ctx)

	if p.awsLambdaRedisPubSubDriver != nil {

		streamPrefix := p.AwsLambdaHandler.ServerStreamTopicPrefix.StringVal()

		p.Log().Info("Starting AWS Lambda stream subscriber", "streamPrefix", streamPrefix)

		p.awsLambdaRedisPubSubDriver.Listen(p.ctx, p.Log().With("component", "awsLambdaRedisPubSubDriver"), true, fmt.Sprintf("%s*", streamPrefix))

	}

}

func (p *GatewayInstance[D]) Stop() {

	p.cancel()

	<-p.gatewayErrCh

	p.ResourceStore.Stop()

	if p.AwsLambdaPubSub != nil {
		p.AwsLambdaPubSub.Close()
	}

}
