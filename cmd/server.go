package main

import (
	"context"
	"strings"

	"dev.azure.com/pomwm/pom-tech/graviflow"
	"dev.azure.com/pomwm/pom-tech/graviflow/internal/client"
	"dev.azure.com/pomwm/pom-tech/graviflow/internal/controlplane"
	apiv1 "dev.azure.com/pomwm/pom-tech/graviflow/proto/api/v1"
	awsprovider "dev.azure.com/pomwm/pom-tech/graviflow/provider/aws"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"google.golang.org/grpc"
)

type resourceStore_Provider string

const (
	dynamoDbResourceStore resourceStore_Provider = "awsdynamodb"
)

type serverDeps interface {
	GetAwsConfig() aws.Config
	GetGrpcServer() *grpc.Server
	GetResourceStoreClient() apiv1.ResourceStoreClient
}

type serverInstance struct {
	graviflow.AppInjector[serverDeps]

	ctx    context.Context
	cancel context.CancelFunc

	resourceStore       *client.GrpcClient[*controllerInstance] `config:"resource.store"`
	resourceStoreClient apiv1.ResourceStoreClient

	resourceStoreProvider graviflow.Config `config:"resource.store.provider,str" usage:"Resource store persistence layer provider"`

	dynamoDbResourceStore *awsprovider.DynamoDBResourceStore[*serverInstance] `config:"dynamodb"`

	enableEnvoyXds graviflow.Config `config:"envoy.xds.enable,bool" usage:"Enable Envoy xDS server"`

	envoyXds      *controlplane.EnvoyXds[*serverInstance] `config:"envoy.xds"`
	envoyXdsErrCh <-chan error
}

func newServerInstance() *serverInstance {
	return &serverInstance{
		resourceStore:         &client.GrpcClient[*controllerInstance]{},
		dynamoDbResourceStore: &awsprovider.DynamoDBResourceStore[*serverInstance]{},
		envoyXds:              &controlplane.EnvoyXds[*serverInstance]{},
	}
}

func (s *serverInstance) GetDynamoDBClient() *dynamodb.Client {
	return dynamodb.NewFromConfig(s.Dependency().GetAwsConfig())
}

func (s *serverInstance) GetGrpcServer() *grpc.Server {
	return s.Dependency().GetGrpcServer()
}

func (s *serverInstance) GetResourceStoreClient() apiv1.ResourceStoreClient {
	return s.resourceStoreClient
}

func (s *serverInstance) Start() {

	s.ctx, s.cancel = context.WithCancel(context.TODO())

	switch resourceStore_Provider(strings.ToLower(s.resourceStoreProvider.StringVal())) {

	case dynamoDbResourceStore:
		s.dynamoDbResourceStore.Initialize()

	}

	s.resourceStore.Start()

	s.resourceStoreClient = apiv1.NewResourceStoreClient(s.resourceStore.ClientConn)

	if s.enableEnvoyXds.BoolVal() {

		s.envoyXds.Initialize()

		s.envoyXdsErrCh = s.envoyXds.Sync(s.ctx)

	}

}

func (s *serverInstance) Stop() {

	s.cancel()

	if s.enableEnvoyXds.BoolVal() {

		<-s.envoyXdsErrCh

	}

	s.resourceStore.Stop()

}
