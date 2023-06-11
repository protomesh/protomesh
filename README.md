# Protomesh

```
Usage of ./protomesh:
  -aws-dynamodb-endpoint-url string
        [string]
                Custom DynamoDB Endpoint url
    
  -aws-enable-grpc-lambda-router
        [boolean]
                Enable gRPC Lambda router
    
  -config-file string
        [string]
                Path to config file (JSON, TOML or YAML)
    
  -enable-envoy-xds
        [boolean]
                Enable envoy xds server instance
    
  -enable-proxy
        [boolean]
                Enable Protomesh proxy instance (synchronized with resource store)
    
  -enable-store
        [boolean]
                Enable Protomesh resource store instance
    
  -enable-worker
        [boolean]
                Enable Protomesh worker instance (synchronized with resource store)
    
  -envoy-xds-resource-store-client-certificate-path string
        [string]
                Path to PEM encoded certificate chain file
    
  -envoy-xds-resource-store-client-certificate-private-key-path string
        [string]
                Path to PEM encoded private key file
    
  -envoy-xds-resource-store-client-enable-tls string
        [string]
                Enable mTLS from client-side
         (default "false")
  -envoy-xds-resource-store-server-address string
        [string]
                gRPC server address to connect to
    
  -envoy-xds-resource-store-server-name-override string
        [string]
                Server name used to verify the hostname returned by TLS handshake
    
  -envoy-xds-server-resource-store-namespace string
        [string]
                Resource store namespace to use
         (default "default")
  -envoy-xds-server-sync-interval duration
        [duration]
                Interval between synchronization cycles
         (default 1m0s)
  -grpc-server-enable-reflection
        [boolean]
                Enable gRPC server reflection
    
  -http-server-shutdown-timeout duration
        [duration]
                HTTP server shutdown timeout before closing
         (default 2m0s)
  -http-server-tls-certificate-certificates-path string
        [string]
                Path to PEM encoded certificate chain file
    
  -http-server-tls-certificate-certificates-private-key-path string
        [string]
                Path to PEM encoded private key file
    
  -http-server-tls-insecure-skip-verify
        [boolean]
                Skip server name verification against certificate chain
    
  -http-server-tls-listener-address string
        [string]
                TLS listener address
    
  -http-server-tls-protocol string
        [string]
                Protocol to accept in the TLS listener
         (default "tcp")
  -http-server-tls-root-cas-path string
        [string]
                Path to PEM encoded certificate chain file
    
  -http-server-tls-root-cas-private-key-path string
        [string]
                Path to PEM encoded private key file
    
  -log-dev
        [boolean]
                Log enhanced for development environment (no sampling)
         (default true)
  -log-json
        [boolean]
                Log in json format
    
  -log-level string
        [string]
                Log level: debug, info or error
    
  -proxy-grpc-router string
        [string]
                Which grpc proxy router to use
    
  -proxy-resource-store-client-certificate-path string
        [string]
                Path to PEM encoded certificate chain file
    
  -proxy-resource-store-client-certificate-private-key-path string
        [string]
                Path to PEM encoded private key file
    
  -proxy-resource-store-client-enable-tls string
        [string]
                Enable mTLS from client-side
         (default "false")
  -proxy-resource-store-server-address string
        [string]
                gRPC server address to connect to
    
  -proxy-resource-store-server-name-override string
        [string]
                Server name used to verify the hostname returned by TLS handshake
    
  -proxy-service-resource-store-namespace string
        [string]
                Resource store namespace to use
         (default "default")
  -proxy-service-sync-interval duration
        [duration]
                Interval between synchronization cycles
         (default 1m0s)
  -store-postgres-migration-file string
        [string]
                Migration file path to execute
    
  -store-postgres-watch-interval duration
        [duration]
                Watch interval between scans
    
  -store-provider string
        [string]
                Resource store persistence layer provider
    
  -store-sql-connection-string string
        [string]
                Connection string to connect to SQL database
    
  -sql-driver-name string
        [string]
                Driver name to use in the SQL client
         (default "postgres")
  -temporal-address string
        [string]
                Tempora server host:port
         (default "localhost:7233")
  -temporal-namespace string
        [string]
                Temporal namespace
         (default "default")
  -worker-resource-store-client-certificate-path string
        [string]
                Path to PEM encoded certificate chain file
    
  -worker-resource-store-client-certificate-private-key-path string
        [string]
                Path to PEM encoded private key file
    
  -worker-resource-store-client-enable-tls string
        [string]
                Enable mTLS from client-side
         (default "false")
  -worker-resource-store-server-address string
        [string]
                gRPC server address to connect to
    
  -worker-resource-store-server-name-override string
        [string]
                Server name used to verify the hostname returned by TLS handshake
    
  -worker-service-resource-store-namespace string
        [string]
                Resource store namespace to use
         (default "default")
  -worker-service-sync-interval duration
        [duration]
                Interval between synchronization cycles
         (default 1m0s)
  -worker-service-worker-task-queue string
        [string]
                Temporal task queue to register activities and workflows
         (default "protomesh")
```

## Building example AWS Lambda

To build the example AWS Lambda that handles the `PingPong.DoPing` gRPC call use the following commands.

```bash
cd examples/lambda
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o main .
rm -rf main.zip
zip main.zip main
```