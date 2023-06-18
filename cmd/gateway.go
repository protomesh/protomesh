package main

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/iancoleman/strcase"
	"github.com/protomesh/go-app"
	"github.com/protomesh/protomesh/pkg/client"
	"github.com/protomesh/protomesh/pkg/gateway"
	servicesv1 "github.com/protomesh/protomesh/proto/api/services/v1"
	awsprovider "github.com/protomesh/protomesh/provider/aws"
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
}

func NewGatewayInstance[D GatewayDeps]() *GatewayInstance[D] {
	return &GatewayInstance[D]{
		ResourceStore:    &client.GrpcClient[GatewayInjector]{},
		Gateway:          &gateway.Gateway[GatewayInjector]{},
		AwsLambdaHandler: &awsprovider.LambdaGatewayHandler[GatewayInjector]{},
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

	handlerTypes := strings.Split(p.Handlers.StringVal(), ",")

	handlers := []gateway.GatewayHandler{}

	for _, handlerType := range handlerTypes {

		switch gateway.HandlerType(strcase.ToSnake(handlerType)) {

		case gateway.HandlerTypeAwsLambda:

			p.AwsLambdaHandler.Initialize()

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
}

func (p *GatewayInstance[D]) Stop() {

	p.cancel()

	<-p.gatewayErrCh

	p.ResourceStore.Stop()

}
