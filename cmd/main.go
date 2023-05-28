package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/upper-institute/graviflow"
	"github.com/upper-institute/graviflow/internal"
	"github.com/upper-institute/graviflow/internal/server"
	awsprovider "github.com/upper-institute/graviflow/provider/aws"
	"github.com/upper-institute/graviflow/provider/temporal"
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

	EnableServer graviflow.Config           `config:"enable.server,bool" default:"false" usage:"Enable Graviflow server instance (resource store and Envoy xDS)"`
	Server       *ServerInstance[*injector] `config:"server"`

	EnableProxy graviflow.Config          `config:"enable.proxy,bool" default:"false" usage:"Enable Graviflow proxy instance (synchronized with resource store)"`
	Proxy       *ProxyInstance[*injector] `config:"proxy"`

	EnableController graviflow.Config               `config:"enable.controller,bool" default:"false" usage:"Enable Graviflow controller instance (temporal worker and service mesh workflows/activities)"`
	Controller       *ControllerInstance[*injector] `config:"controller"`
}

func (i *injector) InjectApp(app graviflow.App) {

	i.Aws = &awsprovider.AwsBuilder[*injector]{}

	i.httpServeMux = http.NewServeMux()

	i.GrpcServer = &server.GrpcServer[*injector]{}

	i.HttpServer = &server.HttpServer[*injector]{
		HttpHandler: i.httpServeMux,
		GrpcHandler: i.GrpcServer,
	}

	i.Temporal = &temporal.TemporalBuilder[*injector]{}

	i.Server = NewServerInstance[*injector]()
	i.Proxy = NewProxyInstance[*injector]()
	i.Controller = NewControllerInstance[*injector]()

	if printConfig {
		graviflow.InjectAppAndPrint(app, i)
		return
	}

	graviflow.InjectApp(app, i)

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

var config = &graviflow.Configurator[*injector]{
	FlagSet:   flag.CommandLine,
	KeyCase:   graviflow.JsonPathCase,
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

	// Initialize AWS SDK config
	dep.Aws.Initialize()

	if dep.EnableProxy.BoolVal() || dep.EnableServer.BoolVal() {
		// Initialize gRPC server
		dep.GrpcServer.Initialize()
	}

	if dep.EnableController.BoolVal() {
		dep.temporalClient = dep.Temporal.GetTemporalClient()
		defer dep.temporalClient.Close()
	}

	if dep.EnableServer.BoolVal() {
		dep.Server.Start()
		defer dep.Server.Stop()
	}

	if dep.EnableProxy.BoolVal() {
		dep.Proxy.Start()
		defer dep.Proxy.Stop()
	}

	if dep.EnableController.BoolVal() {
		dep.Controller.Start()
		defer dep.Controller.Stop()
	}

	if dep.EnableProxy.BoolVal() || dep.EnableServer.BoolVal() {
		// Start and defer stop of gRPC Server
		dep.GrpcServer.Start()
		defer dep.GrpcServer.Start()

		// Start and defer stop of HTTP Server
		dep.HttpServer.Start()
		defer dep.HttpServer.Stop()
	}

	log := app.Log()

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
