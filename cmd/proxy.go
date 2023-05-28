package main

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/upper-institute/graviflow"
	"github.com/upper-institute/graviflow/internal/client"
	"github.com/upper-institute/graviflow/internal/controlplane"
	"github.com/upper-institute/graviflow/internal/server"
	apiv1 "github.com/upper-institute/graviflow/proto/api/v1"
	awsprovider "github.com/upper-institute/graviflow/provider/aws"
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
	GetResourceStoreClient() apiv1.ResourceStoreClient
}

type ProxyInstance[D ProxyDeps] struct {
	*graviflow.AppInjector[D]

	ResourceStore       *client.GrpcClient[ProxyInjector] `config:"resource.store"`
	resourceStoreClient apiv1.ResourceStoreClient

	ctx    context.Context
	cancel context.CancelFunc

	GrpcProxyRouter graviflow.Config `config:"grpc.router,str" usage:"Which grpc proxy router to use"`

	EdgeProxy      *controlplane.EdgeProxy[ProxyInjector] `config:"edge"`
	edgeProxyErrCh <-chan error

	GrpcLambdaRouter *awsprovider.GrpcLambdaRouter[ProxyInjector] `config:"grpc.to.lambda"`
}

func NewProxyInstance[D ProxyDeps]() *ProxyInstance[D] {
	return &ProxyInstance[D]{
		ResourceStore:    &client.GrpcClient[ProxyInjector]{},
		EdgeProxy:        &controlplane.EdgeProxy[ProxyInjector]{},
		GrpcLambdaRouter: &awsprovider.GrpcLambdaRouter[ProxyInjector]{},
	}
}

func (p *ProxyInstance[D]) GetLambdaClient() *lambda.Client {
	return lambda.NewFromConfig(p.Dependency().GetAwsConfig())
}

func (p *ProxyInstance[D]) GetResourceStoreClient() apiv1.ResourceStoreClient {
	return p.resourceStoreClient
}

func (p *ProxyInstance[D]) Start() {

	p.ctx, p.cancel = context.WithCancel(context.TODO())

	p.ResourceStore.Start()

	p.resourceStoreClient = apiv1.NewResourceStoreClient(p.ResourceStore.ClientConn)

	switch grpcProxyRouter(strings.ToLower(p.GrpcProxyRouter.StringVal())) {

	case lambda_grpcProxyRouter:

		p.GrpcLambdaRouter.Initialize()

		p.Dependency().SetGrpcProxyRouter(p.GrpcLambdaRouter)

		p.EdgeProxy.Initialize(p.GrpcLambdaRouter)

	}

	p.edgeProxyErrCh = p.EdgeProxy.Sync(p.ctx)
}

func (p *ProxyInstance[D]) Stop() {

	p.cancel()

	<-p.edgeProxyErrCh

	p.ResourceStore.Stop()

}
