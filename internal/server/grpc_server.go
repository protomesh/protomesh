package server

import (
	"dev.azure.com/pomwm/pom-tech/graviflow"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"

	// "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	// grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GrpcServer[Dependency any] struct {
	graviflow.AppInjector[Dependency]

	*grpc.Server

	GrpcProxy *GrpcProxy

	EnableReflection graviflow.Config `config:"enable.reflection,bool" default:"false" usage:"Enable gRPC server reflection"`
}

func (g *GrpcServer[Dependency]) Initialize() {

	opts := []grpc.ServerOption{
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
			// otelgrpc.StreamServerInterceptor(),
			// grpc_zap.StreamServerInterceptor(internal.Logger),
			),
		),
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
			// otelgrpc.UnaryServerInterceptor(),
			// grpc_zap.UnaryServerInterceptor(internal.Logger),
			),
		),
	}

	if g.GrpcProxy != nil {
		opts = append(
			opts,
			grpc.CustomCodec(Codec()),
			grpc.UnknownServiceHandler(g.GrpcProxy.Handle),
		)
	}

	g.Server = grpc.NewServer(opts...)

	if g.EnableReflection != nil && g.EnableReflection.BoolVal() {
		reflection.Register(g.Server)
	}

}
