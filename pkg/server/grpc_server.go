package server

import (
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/protomesh/go-app"
	"github.com/protomesh/protomesh/pkg/gateway"

	// "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	// grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GrpcGateway interface {
	MatchGrpc(stream grpc.ServerStream) (*gateway.GrpcCall, error)
}

type GrpcServer[Dependency any] struct {
	*app.Injector[Dependency]

	*grpc.Server

	Gateway GrpcGateway

	EnableReflection app.Config `config:"enable.reflection,bool" default:"false" usage:"Enable gRPC server reflection"`
}

func (g *GrpcServer[Dependency]) Initialize() {

	log := g.Log()

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

	if g.Gateway != nil {

		log.Info("Enabling gRPC proxy for unknown services")

		opts = append(
			opts,
			grpc.CustomCodec(Codec()),
			grpc.UnknownServiceHandler(GrpcHandlerFromGateway(g.Gateway)),
		)

	}

	g.Server = grpc.NewServer(opts...)

}

func (g *GrpcServer[Dependency]) Start() {

	log := g.Log()

	if g.EnableReflection != nil && g.EnableReflection.BoolVal() {

		reflection.Register(g.Server)

		log.Info("Reflection registered on gRPC server")

	}

}

func (g *GrpcServer[Dependency]) Stop() {

	g.Server.Stop()

}
