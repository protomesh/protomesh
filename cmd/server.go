package main

import (
	"context"
	"database/sql"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/upper-institute/graviflow"
	"github.com/upper-institute/graviflow/internal/client"
	"github.com/upper-institute/graviflow/internal/controlplane"
	apiv1 "github.com/upper-institute/graviflow/proto/api/v1"
	postgresprovider "github.com/upper-institute/graviflow/provider/postgres"
	"google.golang.org/grpc"
)

type resourceStore_Provider string

const (
	postgresResourceStore resourceStore_Provider = "postgres"
)

type ServerDeps interface {
	GetAwsConfig() aws.Config
	GetGrpcServer() *grpc.Server
}

type ServerInjector interface {
	GetDynamoDBClient() *dynamodb.Client
	GetGrpcServer() *grpc.Server
	GetResourceStoreClient() apiv1.ResourceStoreClient
	GetSqlDatabase() *sql.DB
}

type ServerInstance[D ServerDeps] struct {
	*graviflow.AppInjector[ServerDeps]

	ctx    context.Context
	cancel context.CancelFunc

	ResourceStore       *client.GrpcClient[ServerInjector] `config:"resource.store"`
	resourceStoreClient apiv1.ResourceStoreClient

	ResourceStoreProvider graviflow.Config `config:"resource.store.provider,str" usage:"Resource store persistence layer provider"`

	SqlClient *client.SqlClient[ServerInjector] `config:"sql"`

	PostgresResourceStore *postgresprovider.ResourceStore[ServerInjector] `config:"postgres"`

	EnableEnvoyXds graviflow.Config `config:"envoy.xds.enable,bool" usage:"Enable Envoy xDS server"`

	EnvoyXds      *controlplane.EnvoyXds[ServerInjector] `config:"envoy.xds"`
	envoyXdsErrCh <-chan error
}

func NewServerInstance[D ServerDeps]() *ServerInstance[D] {
	return &ServerInstance[D]{
		ResourceStore:         &client.GrpcClient[ServerInjector]{},
		PostgresResourceStore: &postgresprovider.ResourceStore[ServerInjector]{},
		EnvoyXds:              &controlplane.EnvoyXds[ServerInjector]{},
	}
}

func (s *ServerInstance[D]) GetDynamoDBClient() *dynamodb.Client {
	return dynamodb.NewFromConfig(s.Dependency().GetAwsConfig())
}

func (s *ServerInstance[D]) GetGrpcServer() *grpc.Server {
	return s.Dependency().GetGrpcServer()
}

func (s *ServerInstance[D]) GetResourceStoreClient() apiv1.ResourceStoreClient {
	return s.resourceStoreClient
}

func (s *ServerInstance[D]) GetSqlDatabase() *sql.DB {
	return s.SqlClient.DB
}

func (s *ServerInstance[D]) Start() {

	s.ctx, s.cancel = context.WithCancel(context.TODO())

	log := s.Log()

	provider := strings.ToLower(s.ResourceStoreProvider.StringVal())

	log.Info("Starting Graviflow server", "resourceStoreProvider", provider)

	switch resourceStore_Provider(provider) {

	case postgresResourceStore:

		s.SqlClient.Start()

		s.PostgresResourceStore.Initialize()

	}

	s.ResourceStore.Start()

	s.resourceStoreClient = apiv1.NewResourceStoreClient(s.ResourceStore.ClientConn)

	if s.EnableEnvoyXds.BoolVal() {

		s.EnvoyXds.Initialize()

		s.envoyXdsErrCh = s.EnvoyXds.Sync(s.ctx)

	}

}

func (s *ServerInstance[D]) Stop() {

	s.cancel()

	if s.EnableEnvoyXds.BoolVal() {

		<-s.envoyXdsErrCh

	}

	if s.SqlClient.DB != nil {
		s.SqlClient.Stop()
	}

	s.ResourceStore.Stop()

}
