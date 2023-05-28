package client

import (
	"crypto/x509"

	"github.com/upper-institute/graviflow"

	tlsprovider "github.com/upper-institute/graviflow/provider/tls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcClient[Dependency any] struct {
	*graviflow.AppInjector[Dependency]

	ClientConn *grpc.ClientConn

	TlsBuilder *tlsprovider.CertificateLoader[any] `config:"client.certificate"`
	EnableTls  graviflow.Config                    `config:"client.enable.tls" default:"false" usage:"Enable mTLS from client-side"`

	ServerNameOverride graviflow.Config `config:"server.name.override,str" usage:"Server name used to verify the hostname returned by TLS handshake"`
	ServerAddress      graviflow.Config `config:"server.address,str" usage:"gRPC server address to connect to"`
}

func (g *GrpcClient[Dependency]) Start() {

	log := g.Log()
	addr := g.ServerAddress.StringVal()

	opts := []grpc.DialOption{}

	if g.EnableTls.BoolVal() {

		certs := g.TlsBuilder.BuildDefaultCertificate()

		certPool := x509.NewCertPool()

		for _, cert := range certs {
			certPool.AddCert(cert)
		}

		creds := credentials.NewClientTLSFromCert(certPool, g.ServerNameOverride.StringVal())

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
