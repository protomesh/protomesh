package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	protomesh "github.com/protomesh/protomesh"
	internal "github.com/protomesh/protomesh/pkg"
	"github.com/protomesh/protomesh/pkg/server"
	awsprovider "github.com/protomesh/protomesh/provider/aws"
	"github.com/protomesh/protomesh/provider/temporal"
	temporalcli "go.temporal.io/sdk/client"
	"google.golang.org/grpc"
)

var printConfig = os.Getenv("PRINT_CONFIG") == "true"

type injector struct {
	Aws *awsprovider.AwsBuilder[*injector] `config:"aws"`

	HttpServer   *server.HttpServer[*injector] `config:"http.server"`
	httpServeMux *http.ServeMux

	Temporal       *temporal.TemporalBuilder[*injector] `config:"temporal"`
	temporalClient temporalcli.Client

	GrpcServer *server.GrpcServer[*injector] `config:"grpc.server"`

	EnableStore protomesh.Config `config:"enable.store,bool" default:"false" usage:"Enable Protomesh resource store instance"`
	Store       StoreInjector    `config:"store"`

	EnableEnvoyXds protomesh.Config `config:"enable.envoy.xds,bool" default:"false" usage:"Enable envoy xds server instance"`
	EnvoyXds       EnvoyXdsInjector `config:"envoy.xds"`

	EnableProxy protomesh.Config `config:"enable.proxy,bool" default:"false" usage:"Enable Protomesh proxy instance (synchronized with resource store)"`
	Proxy       ProxyInjector    `config:"proxy"`

	EnableWorker protomesh.Config `config:"enable.worker,bool" default:"false" usage:"Enable Protomesh worker instance (synchronized with resource store)"`
	Worker       WorkerInjector   `config:"worker"`
}

func (i *injector) InjectApp(app protomesh.App) {

	i.Aws = &awsprovider.AwsBuilder[*injector]{}

	i.httpServeMux = http.NewServeMux()

	i.GrpcServer = &server.GrpcServer[*injector]{}

	i.HttpServer = &server.HttpServer[*injector]{
		HttpHandler: i.httpServeMux,
		GrpcHandler: i.GrpcServer,
	}

	i.Temporal = &temporal.TemporalBuilder[*injector]{}

	i.Store = NewStoreInstance[*injector]()
	i.EnvoyXds = NewEnvoyXdsInstance[*injector]()
	i.Proxy = NewProxyInstance[*injector]()
	i.Worker = NewWorkerInstance[*injector]()

	if printConfig {
		protomesh.InjectAndPrint(app, i)
		return
	}

	protomesh.Inject(app, i)

}

func (i *injector) Dependency() *injector {
	return i
}

func (i *injector) GetGrpcServer() *grpc.Server {

	return i.GrpcServer.Server

}

func (i *injector) SetGrpcProxyRouter(router server.GrpcRouter) {

	i.GrpcServer.GrpcProxy = &server.GrpcProxy{
		Router: router,
	}

}

func (i *injector) GetAwsConfig() aws.Config {

	return i.Aws.AwsConfig

}

func (i *injector) GetTemporalClient() temporalcli.Client {

	return i.temporalClient

}

var config = &protomesh.Configurator[*injector]{
	FlagSet:   flag.CommandLine,
	KeyCase:   protomesh.JsonPathCase,
	Separator: ".",
	Print:     printConfig,
}

func main() {

	dep := &injector{}

	// Apply configuration for dependencies
	config.ApplyFlags(dep)

	// Parse command flags
	flag.Parse()

	// Create app with dependency set and configurator
	app := internal.CreateApp[*injector](dep, config)
	defer app.Close()

	// Inject app into dependency set
	dep.InjectApp(app)

	// Apply loaded configurations to dependency set
	config.ApplyConfigs(dep)

	log := app.Log()

	// Initialize AWS SDK config
	dep.Aws.Initialize()

	if dep.EnableProxy.BoolVal() {
		dep.Proxy.Initialize()
		log.Info("Proxy initialized")
	}

	if dep.EnableProxy.BoolVal() || dep.EnableStore.BoolVal() || dep.EnableEnvoyXds.BoolVal() {
		// Initialize gRPC server
		dep.GrpcServer.Initialize()
		log.Info("gRPC server initialized")
	}

	if dep.EnableWorker.BoolVal() {
		dep.temporalClient = dep.Temporal.GetTemporalClient()
		defer dep.temporalClient.Close()
	}

	if dep.EnableStore.BoolVal() {
		dep.Store.Start()
		log.Info("Resource store started")
		defer dep.Store.Stop()
	}

	if dep.EnableProxy.BoolVal() {
		dep.Proxy.Start()
		log.Info("Proxy started")
		defer dep.Proxy.Stop()
	}

	if dep.EnableEnvoyXds.BoolVal() {
		dep.EnvoyXds.Start()
		log.Info("Envoy xDS server started")
		defer dep.EnvoyXds.Stop()
	}

	if dep.EnableWorker.BoolVal() {
		dep.Worker.Start()
		defer dep.Worker.Stop()
	}

	if dep.EnableProxy.BoolVal() || dep.EnableStore.BoolVal() || dep.EnableEnvoyXds.BoolVal() {

		// Start and defer stop of gRPC Server
		dep.GrpcServer.Start()
		log.Info("gRPC server started")
		defer dep.GrpcServer.Start()

		// Start and defer stop of HTTP Server
		dep.HttpServer.Start()
		log.Info("HTTP server started")
		defer dep.HttpServer.Stop()
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Panic("Failed to get hostname", "error", err)
	}

	pid := os.Getpid()
	uid := os.Getuid()

	log.Info("Application started", "hostname", hostname, "pid", pid, "uid", uid)

	// Wait for interruption signal
	internal.WaitInterruption()

}
