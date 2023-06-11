package main

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/protomesh/protomesh"
	"github.com/protomesh/protomesh/pkg/client"
	"github.com/protomesh/protomesh/pkg/proxy"
	"github.com/protomesh/protomesh/pkg/server"
	servicesv1 "github.com/protomesh/protomesh/proto/api/services/v1"
	awsprovider "github.com/protomesh/protomesh/provider/aws"
	"google.golang.org/grpc"
)

type grpcProxyRouter string

const (
	lambda_grpcProxyRouter grpcProxyRouter = "awslambda"
)

type ProxyDeps interface {
	GetAwsConfig() aws.Config
	GetGrpcServer() *grpc.Server
	SetGrpcProxyRouter(router server.GrpcRouter)
}

type ProxyInjector interface {
	GetLambdaClient() *lambda.Client
	GetResourceStoreClient() servicesv1.ResourceStoreClient

	Initialize()
	Start()
	Stop()
}

type ProxyInstance[D ProxyDeps] struct {
	*protomesh.Injector[D]

	ResourceStore       *client.GrpcClient[ProxyInjector] `config:"resource.store"`
	resourceStoreClient servicesv1.ResourceStoreClient

	ctx    context.Context
	cancel context.CancelFunc

	GrpcProxyRouter protomesh.Config `config:"grpc.router,str" usage:"Which grpc proxy router to use"`

	Proxy          *proxy.Proxy[ProxyInjector] `config:"service"`
	edgeProxyErrCh <-chan error

	GrpcLambdaRouter *awsprovider.GrpcLambdaRouter[ProxyInjector] `config:"grpc.to.lambda"`
}

func NewProxyInstance[D ProxyDeps]() ProxyInjector {
	return &ProxyInstance[D]{
		ResourceStore:    &client.GrpcClient[ProxyInjector]{},
		Proxy:            &proxy.Proxy[ProxyInjector]{},
		GrpcLambdaRouter: &awsprovider.GrpcLambdaRouter[ProxyInjector]{},
	}
}

func (p *ProxyInstance[D]) GetLambdaClient() *lambda.Client {
	return lambda.NewFromConfig(p.Dependency().GetAwsConfig())
}

func (p *ProxyInstance[D]) GetResourceStoreClient() servicesv1.ResourceStoreClient {
	return p.resourceStoreClient
}

func (p *ProxyInstance[D]) Initialize() {

	log := p.Log()

	compRouter := server.NewCompositeGrpcRouter()
	handlers := []proxy.ProxyHandler{}

	grpcProxyRouterStr := strings.ToLower(p.GrpcProxyRouter.StringVal())

	switch grpcProxyRouter(grpcProxyRouterStr) {

	case lambda_grpcProxyRouter:

		p.GrpcLambdaRouter.Initialize()

		compRouter = append(compRouter, p.GrpcLambdaRouter)
		handlers = append(handlers, p.GrpcLambdaRouter)

		log.Info("Initialized gRPC proxy router using gRPC Lambda Router (AWS)")

	default:
		log.Panic("Unknown gRPC proxy router", "grpcProxyRouter", grpcProxyRouterStr)

	}

	p.Dependency().SetGrpcProxyRouter(compRouter)
	p.Proxy.Initialize(handlers...)

}

func (p *ProxyInstance[D]) Start() {

	p.ctx, p.cancel = context.WithCancel(context.TODO())

	p.ResourceStore.Start()

	p.resourceStoreClient = servicesv1.NewResourceStoreClient(p.ResourceStore.ClientConn)

	p.edgeProxyErrCh = p.Proxy.Sync(p.ctx)
}

func (p *ProxyInstance[D]) Stop() {

	p.cancel()

	<-p.edgeProxyErrCh

	p.ResourceStore.Stop()

}
