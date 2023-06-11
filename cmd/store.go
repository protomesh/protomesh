package main

import (
	"database/sql"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/protomesh/protomesh"
	"github.com/protomesh/protomesh/pkg/client"
	"github.com/protomesh/protomesh/provider/postgres"
	"google.golang.org/grpc"
)

type resourceStoreProvider string

const (
	postgresResourceStore resourceStoreProvider = "postgres"
)

type StoreDeps interface {
	GetAwsConfig() aws.Config
	GetGrpcServer() *grpc.Server
}

type StoreInjector interface {
	GetDynamoDBClient() *dynamodb.Client
	GetGrpcServer() *grpc.Server
	GetSqlDatabase() *sql.DB

	Start()
	Stop()
}

type StoreInstance[D StoreDeps] struct {
	*protomesh.Injector[D]

	ResourceStoreProvider protomesh.Config `config:"provider,str" usage:"Resource store persistence layer provider"`

	SqlClient *client.SqlClient[StoreInjector] `config:"sql"`

	PostgresResourceStore *postgres.ResourceStore[StoreInjector] `config:"postgres"`
}

func NewStoreInstance[D StoreDeps]() StoreInjector {
	return &StoreInstance[D]{
		PostgresResourceStore: &postgres.ResourceStore[StoreInjector]{},
	}
}

func (s *StoreInstance[D]) GetDynamoDBClient() *dynamodb.Client {
	return dynamodb.NewFromConfig(s.Dependency().GetAwsConfig())
}

func (s *StoreInstance[D]) GetGrpcServer() *grpc.Server {
	return s.Dependency().GetGrpcServer()
}

func (s *StoreInstance[D]) GetSqlDatabase() *sql.DB {
	return s.SqlClient.DB
}

func (s *StoreInstance[D]) Start() {

	log := s.Log()

	provider := strings.ToLower(s.ResourceStoreProvider.StringVal())

	log.Info("Starting Protomesh resource store", "resourceStoreProvider", provider)

	switch resourceStoreProvider(provider) {

	case postgresResourceStore:

		s.SqlClient.Start()

		s.PostgresResourceStore.Initialize()

	}

}

func (s *StoreInstance[D]) Stop() {

	if s.SqlClient.DB != nil {
		s.SqlClient.Stop()
	}

}
