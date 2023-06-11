package main

import (
	"context"

	protomesh "github.com/protomesh/protomesh"
	"github.com/protomesh/protomesh/pkg/client"
	servicesv1 "github.com/protomesh/protomesh/proto/api/services/v1"
	"github.com/protomesh/protomesh/provider/envoy"
	"google.golang.org/grpc"
)

type EnvoyXdsDeps interface {
	GetGrpcServer() *grpc.Server
}

type EnvoyXdsInjector interface {
	GetGrpcServer() *grpc.Server
	GetResourceStoreClient() servicesv1.ResourceStoreClient

	Start()
	Stop()
}

type EnvoyXdsInstance[D EnvoyXdsDeps] struct {
	*protomesh.Injector[D]

	ctx    context.Context
	cancel context.CancelFunc

	ResourceStore       *client.GrpcClient[EnvoyXdsInjector] `config:"resource.store"`
	resourceStoreClient servicesv1.ResourceStoreClient

	EnvoyXds      *envoy.EnvoyXds[EnvoyXdsInjector] `config:"server"`
	envoyXdsErrCh <-chan error
}

func NewEnvoyXdsInstance[D EnvoyXdsDeps]() EnvoyXdsInjector {
	return &EnvoyXdsInstance[D]{
		ResourceStore: &client.GrpcClient[EnvoyXdsInjector]{},
		EnvoyXds:      &envoy.EnvoyXds[EnvoyXdsInjector]{},
	}
}

func (e *EnvoyXdsInstance[D]) GetResourceStoreClient() servicesv1.ResourceStoreClient {
	return e.resourceStoreClient
}

func (e *EnvoyXdsInstance[D]) GetGrpcServer() *grpc.Server {
	return e.Dependency().GetGrpcServer()
}

func (e *EnvoyXdsInstance[D]) Start() {

	e.ctx, e.cancel = context.WithCancel(context.TODO())

	e.ResourceStore.Start()

	e.resourceStoreClient = servicesv1.NewResourceStoreClient(e.ResourceStore.ClientConn)

	e.EnvoyXds.Initialize()

	e.envoyXdsErrCh = e.EnvoyXds.Sync(e.ctx)

}

func (e *EnvoyXdsInstance[D]) Stop() {

	e.cancel()

	<-e.envoyXdsErrCh

	e.ResourceStore.Stop()

}
