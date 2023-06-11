package main

import (
	"context"
	"errors"
	"os"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/protomesh/protomesh"
	"github.com/protomesh/protomesh/pkg/client"
	"github.com/protomesh/protomesh/pkg/config"
	servicesv1 "github.com/protomesh/protomesh/proto/api/services/v1"
	"github.com/protomesh/protomesh/provider/tls"
)

func kvIntoMap(kv []interface{}) map[string]interface{} {

	m := make(map[string]interface{})

	var key string

	for i, v := range kv {

		if i%2 == 0 {
			key = v.(string)
			continue
		}

		m[key] = v

	}

	return m

}

type providerLogger struct {
	ctx  context.Context
	name string
	kv   []interface{}
}

func (p *providerLogger) addNameToKv(kv []interface{}) []interface{} {
	kv = append(kv, p.kv...)
	if len(p.name) > 0 {
		kv = append(kv, "name", p.name)
	}
	return kv
}

func (p *providerLogger) Debug(message string, kv ...interface{}) {
	tflog.Debug(p.ctx, message, kvIntoMap(p.addNameToKv(kv)))
}

func (p *providerLogger) Info(message string, kv ...interface{}) {
	tflog.Info(p.ctx, message, kvIntoMap(p.addNameToKv(kv)))
}

func (p *providerLogger) Warn(message string, kv ...interface{}) {
	tflog.Warn(p.ctx, message, kvIntoMap(p.addNameToKv(kv)))
}

func (p *providerLogger) Error(message string, kv ...interface{}) {
	tflog.Error(p.ctx, message, kvIntoMap(p.addNameToKv(kv)))
}

func (p *providerLogger) Panic(message string, kv ...interface{}) {
	tflog.Error(p.ctx, message, kvIntoMap(p.addNameToKv(kv)))
	os.Exit(1)
}

func (p *providerLogger) With(kv ...interface{}) protomesh.Logger {
	return &providerLogger{
		ctx:  p.ctx,
		name: p.name,
		kv:   kv,
	}
}

type providerApp[D any] struct {
	log *providerLogger
}

func (a *providerApp[D]) Config() protomesh.ConfigSource {
	return nil
}

func (a *providerApp[D]) Log() protomesh.Logger {
	return a.log
}

func (a *providerApp[D]) Close() {
}

type providerDeps struct {
	GrpcClient *client.GrpcClient[any]
}

func (d *providerDeps) Dependency() *providerDeps {
	return d
}

func (d *providerDeps) GetResourceStoreClient() servicesv1.ResourceStoreClient {
	return servicesv1.NewResourceStoreClient(d.GrpcClient.ClientConn)
}

type dependencies interface {
	GetResourceStoreClient() servicesv1.ResourceStoreClient
}

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"address": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PROTOMESH_SERVER_ADDRESS", "localhost:6680"),
			},
			"enable_tls": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PROTOMESH_ENABLE_TLS", false),
			},
			"server_name_override": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PROTOMESH_SERVER_NAME_OVERRIDE", nil),
			},
			"tls_certificate_path": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PROTOMESH_TLS_CERTIFICATE_PATH", nil),
			},
			"tls_private_key_path": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PROTOMESH_TLS_PRIVATE_KEY_PATH", nil),
			},
		},
		ConfigureContextFunc: func(ctx context.Context, rd *schema.ResourceData) (interface{}, diag.Diagnostics) {

			tflog.Info(ctx, "Configuring HashiCups client")

			address, ok := rd.Get("address").(string)
			if !ok {
				return nil, diag.FromErr(errors.New("Invalid address type provided"))
			}

			enableTls, ok := rd.Get("enable_tls").(bool)
			if !ok {
				return nil, diag.FromErr(errors.New("Invalid enable_tls type provided."))
			}

			deps := &providerDeps{
				GrpcClient: &client.GrpcClient[any]{
					EnableTls: config.NewConfig(enableTls),
					TlsBuilder: &tls.CertificateLoader[any]{
						CertificatePath: config.EmptyConfig(),
						PrivateKey: &tls.KeyLoader[any]{
							KeysPath: config.EmptyConfig(),
						},
					},
					ServerAddress:      config.NewConfig(address),
					ServerNameOverride: config.EmptyConfig(),
				},
			}

			serverNameOverride := rd.Get("server_name_override")
			if serverNameOverride != nil {
				deps.GrpcClient.ServerNameOverride = config.NewConfig(serverNameOverride)
			}

			tlsCertPath := rd.Get("tls_certificate_path")
			if tlsCertPath != nil {
				deps.GrpcClient.TlsBuilder.CertificatePath = config.NewConfig(tlsCertPath)
			}

			tlsPrivPath := rd.Get("tls_private_key_path")
			if tlsPrivPath != nil {
				deps.GrpcClient.TlsBuilder.PrivateKey.KeysPath = config.NewConfig(tlsPrivPath)
			}

			app := &providerApp[*providerDeps]{
				log: &providerLogger{
					ctx:  ctx,
					name: "protomesh",
					kv:   make([]interface{}, 0),
				},
			}

			protomesh.Inject(app, deps)

			deps.GrpcClient.Start()

			return deps, nil
		},
		ResourcesMap: map[string]*schema.Resource{
			"protomesh_aws_lambda_grpc": resourceAwsLambdaGrpc(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"protomesh_aws_lambda_grpc": dataSourceAwsLambdaGrpc(),
		},
	}
}
