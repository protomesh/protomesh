package main

import (
	"context"
	"strings"

	"dev.azure.com/pomwm/pom-tech/graviflow"
	"dev.azure.com/pomwm/pom-tech/graviflow/internal/client"
	"dev.azure.com/pomwm/pom-tech/graviflow/internal/controlplane"
	"dev.azure.com/pomwm/pom-tech/graviflow/internal/server"
	apiv1 "dev.azure.com/pomwm/pom-tech/graviflow/proto/api/v1"
	awsprovider "dev.azure.com/pomwm/pom-tech/graviflow/provider/aws"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"google.golang.org/grpc"
)

type grpcProxyRouter string

const (
	lambda_grpcProxyRouter grpcProxyRouter = "awslambda"
)

type proxyDeps interface {
	GetAwsConfig() aws.Config
	GetGrpcServer() *grpc.Server
	SetGrpcProxyRouter(router server.GrpcRouter)
}

type proxyInstance struct {
	graviflow.AppInjector[proxyDeps]

	resourceStore       *client.GrpcClient[*controllerInstance] `config:"resource.store"`
	resourceStoreClient apiv1.ResourceStoreClient

	ctx    context.Context
	cancel context.CancelFunc

	grpcProxyRouter graviflow.Config `config:"grpc.router,str" usage:"Which grpc proxy router to use"`

	edgeProxy      *controlplane.EdgeProxy[*proxyInstance] `config:"edge"`
	edgeProxyErrCh <-chan error

	grpcLambdaRouter *awsprovider.GrpcLambdaRouter[*proxyInstance] `config:"grpc.to.lambda"`
}

func newProxyInstance() *proxyInstance {
	return &proxyInstance{
		resourceStore:    &client.GrpcClient[*controllerInstance]{},
		edgeProxy:        &controlplane.EdgeProxy[*proxyInstance]{},
		grpcLambdaRouter: &awsprovider.GrpcLambdaRouter[*proxyInstance]{},
	}
}

func (p *proxyInstance) GetLambdaClient() *lambda.Client {
	return lambda.NewFromConfig(p.Dependency().GetAwsConfig())
}

func (p *proxyInstance) GetResourceStoreClient() apiv1.ResourceStoreClient {
	return p.resourceStoreClient
}

func (p *proxyInstance) Start() {

	p.ctx, p.cancel = context.WithCancel(context.TODO())

	p.resourceStore.Start()

	p.resourceStoreClient = apiv1.NewResourceStoreClient(p.resourceStore.ClientConn)

	switch grpcProxyRouter(strings.ToLower(p.grpcProxyRouter.StringVal())) {

	case lambda_grpcProxyRouter:

		p.grpcLambdaRouter.Initialize()

		p.Dependency().SetGrpcProxyRouter(p.grpcLambdaRouter)

		p.edgeProxy.Initialize(p.grpcLambdaRouter)

	}

	p.edgeProxyErrCh = p.edgeProxy.Sync(p.ctx)
}

func (p *proxyInstance) Stop() {

	p.cancel()

	<-p.edgeProxyErrCh

	p.resourceStore.Stop()

}
