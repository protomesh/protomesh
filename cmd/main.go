package main

import (
	"flag"
	"net/http"
	"os"

	"dev.azure.com/pomwm/pom-tech/graviflow"
	"dev.azure.com/pomwm/pom-tech/graviflow/internal"
	"dev.azure.com/pomwm/pom-tech/graviflow/internal/server"
	awsprovider "dev.azure.com/pomwm/pom-tech/graviflow/provider/aws"
	"dev.azure.com/pomwm/pom-tech/graviflow/provider/temporal"
	tlsprovider "dev.azure.com/pomwm/pom-tech/graviflow/provider/tls"
	"github.com/aws/aws-sdk-go-v2/aws"
	temporalcli "go.temporal.io/sdk/client"
	"google.golang.org/grpc"
)

type injector struct {
	aws *awsprovider.AwsBuilder[*injector] `config:"aws"`

	httpServer   *server.HttpServer[*injector] `config:"http.server"`
	httpServeMux *http.ServeMux

	temporal       *temporal.TemporalBuilder[*injector] `config:"temporal"`
	temporalClient temporalcli.Client

	grpcServer *server.GrpcServer[*injector] `config:"grpc.server"`

	enableServer graviflow.Config `config:"enable.server,bool" default:"false" usage:"Enable Graviflow server instance (resource store and Envoy xDS)"`
	server       *serverInstance  `config:"server"`

	enableProxy graviflow.Config `config:"enable.proxy,bool" default:"false" usage:"Enable Graviflow proxy instance (synchronized with resource store)"`
	proxy       *proxyInstance   `config:"proxy"`

	enableController graviflow.Config    `config:"enable.controller,bool" default:"false" usage:"Enable Graviflow controller instance (temporal worker and service mesh workflows/activities)"`
	controller       *controllerInstance `config:"controller"`
}

func (i *injector) InjectApp(app graviflow.App[*injector]) {

	i.aws = &awsprovider.AwsBuilder[*injector]{}

	i.httpServeMux = http.NewServeMux()

	i.grpcServer = &server.GrpcServer[*injector]{}

	i.httpServer = &server.HttpServer[*injector]{
		TlsBuilder: &tlsprovider.TlsBuilder[*injector]{
			Certificate: &tlsprovider.TlsCertificateLoader[*injector]{
				PrivateKey:   &tlsprovider.KeyLoader[*injector]{},
				Certificates: &tlsprovider.CertificateLoader[*injector]{},
			},
			RootCAs: &tlsprovider.CertificateLoader[*injector]{},
		},
		HttpHandler: i.httpServeMux,
		GrpcHandler: i.grpcServer,
	}

	i.temporal = &temporal.TemporalBuilder[*injector]{}

	i.server = newServerInstance()
	i.proxy = newProxyInstance()
	i.controller = newControllerInstance()

	graviflow.InjectApp(app, i)

}

func (i *injector) GetGrpcServer() *grpc.Server {

	return i.grpcServer.Server

}

func (i *injector) SetGrpcProxyRouter(router server.GrpcRouter) {

	i.grpcServer.GrpcProxy = &server.GrpcProxy{
		Router: router,
	}

}

func (i *injector) GetAwsConfig() aws.Config {

	return i.aws.Config

}

func (i *injector) GetTemporalClient() temporalcli.Client {

	return i.temporalClient

}

func (i *injector) Dependency() *injector {
	return i
}

var config = &graviflow.Configurator[*injector]{
	FlagSet:   flag.CommandLine,
	KeyCase:   graviflow.SnakeCaseKey,
	Separator: ".",
	Print:     os.Getenv("PRINT_CONFIG") == "true",
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
	dep.aws.Initialize()

	if dep.enableProxy.BoolVal() || dep.enableServer.BoolVal() {
		// Initialize gRPC server
		dep.grpcServer.Initialize()

		// Start and defer stop of HTTP Server
		dep.httpServer.Start()
		defer dep.httpServer.Stop()
	}

	if dep.enableController.BoolVal() {
		dep.temporalClient = dep.temporal.GetTemporalClient()
		defer dep.temporalClient.Close()
	}

	if dep.enableServer.BoolVal() {
		dep.server.Start()
		defer dep.server.Stop()
	}

	if dep.enableProxy.BoolVal() {
		dep.proxy.Start()
		defer dep.proxy.Stop()
	}

	if dep.enableController.BoolVal() {
		dep.controller.Start()
		defer dep.controller.Stop()
	}

	// Wait for interruption signal
	internal.WaitInterruption()

}
