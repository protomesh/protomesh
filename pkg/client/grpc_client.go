package client

import (
	"crypto/tls"
	"crypto/x509"

	"github.com/protomesh/go-app"
	tlsprovider "github.com/protomesh/protomesh/provider/tls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcClient[Dependency any] struct {
	*app.Injector[Dependency]

	ClientConn *grpc.ClientConn

	TlsBuilder *tlsprovider.CertificateLoader[any] `config:"client.certificate"`
	EnableTls  app.Config                          `config:"client.enable.tls" default:"false" usage:"Enable mTLS from client-side"`

	ServerNameOverride app.Config `config:"server.name.override,str" usage:"Server name used to verify the hostname returned by TLS handshake"`
	ServerAddress      app.Config `config:"server.address,str" usage:"gRPC server address to connect to"`
}

func (g *GrpcClient[Dependency]) Start() {

	log := g.Log()
	addr := g.ServerAddress.StringVal()

	opts := []grpc.DialOption{
		grpc.WithDefaultServiceConfig(`{
			"loadBalancingConfig": [ { "round_robin": {} } ],
			"methodConfig": [{
				"name": [{"service": "protomesh.api.v1.ResourceStore"}],
				"waitForReady": true,

				"retryPolicy": {
					"MaxAttempts": 5,
					"InitialBackoff": "5s",
					"MaxBackoff": "30s",
					"BackoffMultiplier": 1.0,
					"RetryableStatusCodes": [ "UNAVAILABLE" ]
				}
			}]
		}`),
	}

	if g.EnableTls.BoolVal() {

		tlsCfg := &tls.Config{
			RootCAs:            x509.NewCertPool(),
			InsecureSkipVerify: true,
		}

		certs := g.TlsBuilder.BuildCertificates()

		for _, cert := range certs {
			tlsCfg.RootCAs.AddCert(cert)
		}

		creds := credentials.NewTLS(tlsCfg)

		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	clientConn, err := grpc.Dial(addr, opts...)
	if err != nil {
		log.Panic("Error dialing gRPC server", "error", err, "address", addr)
	}

	g.ClientConn = clientConn

}

func (g *GrpcClient[Dependency]) Stop() {

	log := g.Log()

	err := g.ClientConn.Close()
	if err != nil {
		log.Panic("Error while stopping gRPC client conn", "error", err)
	}

}
