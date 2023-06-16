package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/protomesh/go-app"
	"github.com/protomesh/protomesh/pkg/server"
	awsprovider "github.com/protomesh/protomesh/provider/aws"
	"github.com/protomesh/protomesh/provider/temporal"
	temporalcli "go.temporal.io/sdk/client"
	"google.golang.org/grpc"
)

type root struct {
	*app.Injector[*root]

	Aws *awsprovider.AwsBuilder[*root] `config:"aws"`

	HttpServer   *server.HttpServer[*root] `config:"http.server"`
	httpServeMux *http.ServeMux

	Temporal       *temporal.TemporalBuilder[*root] `config:"temporal"`
	temporalClient temporalcli.Client

	GrpcServer *server.GrpcServer[*root] `config:"grpc.server"`

	EnableStore app.Config            `config:"enable.store,bool" default:"false" usage:"Enable Protomesh resource store instance"`
	Store       *StoreInstance[*root] `config:"store"`

	EnableEnvoyXds app.Config               `config:"enable.envoy.xds,bool" default:"false" usage:"Enable envoy xds server instance"`
	EnvoyXds       *EnvoyXdsInstance[*root] `config:"envoy.xds"`

	EnableGateway app.Config              `config:"enable.gateway,bool" default:"false" usage:"Enable Protomesh gateway instance (synchronized with resource store)"`
	Gateway       *GatewayInstance[*root] `config:"gateway"`

	EnableWorker app.Config             `config:"enable.worker,bool" default:"false" usage:"Enable Protomesh worker instance (synchronized with resource store)"`
	Worker       *WorkerInstance[*root] `config:"worker"`
}

func newRoot() *root {

	grpcServer := &server.GrpcServer[*root]{}

	return &root{
		Aws: &awsprovider.AwsBuilder[*root]{},
		HttpServer: &server.HttpServer[*root]{
			HttpHandler: http.NewServeMux(),
			GrpcHandler: grpcServer,
		},
		httpServeMux: http.NewServeMux(),
		Temporal:     &temporal.TemporalBuilder[*root]{},
		GrpcServer:   grpcServer,
		Store:        NewStoreInstance[*root](),
		EnvoyXds:     NewEnvoyXdsInstance[*root](),
		Gateway:      NewGatewayInstance[*root](),
		Worker:       NewWorkerInstance[*root](),
	}

}

func (i *root) Dependency() *root {
	return i
}

func (i *root) GetGrpcServer() *grpc.Server {
	return i.GrpcServer.Server
}

func (i *root) SetGrpcProxyRouter(router server.GrpcRouter) {
	i.GrpcServer.GrpcProxy = &server.GrpcProxy{
		Router: router,
	}
}

func (i *root) GetAwsConfig() aws.Config {
	return i.Aws.AwsConfig
}

func (i *root) GetTemporalClient() temporalcli.Client {
	return i.temporalClient
}

var opts = &app.AppOptions{
	FlagSet: flag.CommandLine,
	Print:   os.Getenv("PRINT_CONFIG") == "true",
}

func main() {

	deps := newRoot()

	cmdApp := app.NewApp(deps, opts)
	defer cmdApp.Close()

	log := cmdApp.Log()

	// Initialize AWS SDK config
	deps.Aws.Initialize()

	if deps.EnableGateway.BoolVal() {
		deps.Gateway.Initialize()
		log.Info("Gateway initialized")
	}

	if deps.EnableGateway.BoolVal() || deps.EnableStore.BoolVal() || deps.EnableEnvoyXds.BoolVal() {
		// Initialize gRPC server
		deps.GrpcServer.Initialize()
		log.Info("gRPC server initialized")
	}

	if deps.EnableWorker.BoolVal() {
		deps.temporalClient = deps.Temporal.GetTemporalClient()
		defer deps.temporalClient.Close()
	}

	if deps.EnableStore.BoolVal() {
		deps.Store.Start()
		log.Info("Resource store started")
		defer deps.Store.Stop()
	}

	if deps.EnableGateway.BoolVal() {
		deps.Gateway.Start()
		log.Info("Gateway started")
		defer deps.Gateway.Stop()
	}

	if deps.EnableEnvoyXds.BoolVal() {
		deps.EnvoyXds.Start()
		log.Info("Envoy xDS server started")
		defer deps.EnvoyXds.Stop()
	}

	if deps.EnableWorker.BoolVal() {
		deps.Worker.Start()
		defer deps.Worker.Stop()
	}

	if deps.EnableGateway.BoolVal() || deps.EnableStore.BoolVal() || deps.EnableEnvoyXds.BoolVal() {

		// Start and defer stop of gRPC Server
		deps.GrpcServer.Start()
		log.Info("gRPC server started")
		defer deps.GrpcServer.Start()

		// Start and defer stop of HTTP Server
		deps.HttpServer.Start()
		log.Info("HTTP server started")
		defer deps.HttpServer.Stop()
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Panic("Failed to get hostname", "error", err)
	}

	pid := os.Getpid()
	uid := os.Getuid()

	log.Info("Application started", "hostname", hostname, "pid", pid, "uid", uid)

	// Wait for interruption signal
	app.WaitInterruption()

}
