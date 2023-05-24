package controlplane

import (
	apiv1 "dev.azure.com/pomwm/pom-tech/graviflow/proto/api/v1"
	"go.temporal.io/sdk/client"
	"google.golang.org/grpc"
)

type TemporalProvider interface {
	GetTemporalClient() client.Client
}

type GrpcServerProvider interface {
	GetGrpcServer() *grpc.Server
}

type ResourceStoreProvider interface {
	GetResourceStoreClient() apiv1.ResourceStoreClient
}
