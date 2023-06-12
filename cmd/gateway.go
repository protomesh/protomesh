package main

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/protomesh/protomesh"
	"github.com/protomesh/protomesh/pkg/client"
	"github.com/protomesh/protomesh/pkg/gateway"
	"github.com/protomesh/protomesh/pkg/server"
	servicesv1 "github.com/protomesh/protomesh/proto/api/services/v1"
	awsprovider "github.com/protomesh/protomesh/provider/aws"
	"google.golang.org/grpc"
)

type grpcGatewayRouter string

const (
	lambda_grpcGatewayRouter grpcGatewayRouter = "awslambda"
)

type GatewayDeps interface {
	GetAwsConfig() aws.Config
	GetGrpcServer() *grpc.Server
	SetGrpcProxyRouter(router server.GrpcRouter)
}

type GatewayInjector interface {
	GetLambdaClient() *lambda.Client
	GetResourceStoreClient() servicesv1.ResourceStoreClient

	Initialize()
	Start()
	Stop()
}

type GatewayInstance[D GatewayDeps] struct {
	*protomesh.Injector[D]

	ResourceStore       *client.GrpcClient[GatewayInjector] `config:"resource.store"`
	resourceStoreClient servicesv1.ResourceStoreClient

	ctx    context.Context
	cancel context.CancelFunc

	GrpcProxyRouter protomesh.Config `config:"grpc.router,str" usage:"Which grpc proxy router to use"`

	Gateway      *gateway.Gateway[GatewayInjector] `config:"service"`
	gatewayErrCh <-chan error

	GrpcLambdaRouter *awsprovider.GrpcLambdaRouter[GatewayInjector] `config:"grpc.to.lambda"`
}

func NewGatewayInstance[D GatewayDeps]() *GatewayInstance[D] {
	return &GatewayInstance[D]{
		ResourceStore:    &client.GrpcClient[GatewayInjector]{},
		Gateway:          &gateway.Gateway[GatewayInjector]{},
		GrpcLambdaRouter: &awsprovider.GrpcLambdaRouter[GatewayInjector]{},
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

	compRouter := server.NewCompositeGrpcRouter()
	handlers := []gateway.GatewayHandler{}

	grpcGatewayRouterStr := strings.ToLower(p.GrpcProxyRouter.StringVal())

	switch grpcGatewayRouter(grpcGatewayRouterStr) {

	case lambda_grpcGatewayRouter:

		p.GrpcLambdaRouter.Initialize()

		compRouter = append(compRouter, p.GrpcLambdaRouter)
		handlers = append(handlers, p.GrpcLambdaRouter)

		log.Info("Initialized gRPC proxy router using gRPC Lambda Router (AWS)")

	default:
		log.Panic("Unknown gRPC proxy router", "grpcProxyRouter", grpcGatewayRouterStr)

	}

	p.Dependency().SetGrpcProxyRouter(compRouter)
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
