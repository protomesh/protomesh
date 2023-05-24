# Introduction 
TODO: Give a short introduction of your project. Let this section explain the objectives or the motivation behind this project. 

# Getting Started
TODO: Guide users through getting your code up and running on their own system. In this section you can talk about:
1.	Installation process
2.	Software dependencies
3.	Latest releases
4.	API references

# Build and Test
TODO: Describe and show how to build your code and run the tests. 

# Contribute
TODO: Explain how other users and developers can contribute to make your code better. 

If you want to learn more about creating good readme files then refer the following [guidelines](https://docs.microsoft.com/en-us/azure/devops/repos/git/create-a-readme?view=azure-devops). You can also seek inspiration from the below readme files:
- [ASP.NET Core](https://github.com/aspnet/Home)
- [Visual Studio Code](https://github.com/Microsoft/vscode)
- [Chakra Core](https://github.com/Microsoft/ChakraCore)


-aws-enable-grpc-lambda-router
        [boolean]
                Enable gRPC Lambda router
    
  -config-file string
        [string]
                Path to config file (JSON, TOML or YAML)
    
  -controller-controls-worker-task-queue string
        [string]
                Temporal task queue to register activities and workflows
         (default "graviflow")
  -enable-controller
        [boolean]
                Enable Graviflow controller instance (temporal worker and service mesh workflows/activities)
    
  -enable-proxy
        [boolean]
                Enable Graviflow proxy instance (synchronized with resource store)
    
  -enable-server
        [boolean]
                Enable Graviflow server instance (resource store and Envoy xDS)
    
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
    
  -http-server-tls-certificate-private-key-path string
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
    
  -log-dev
        [boolean]
                Log enhanced for development environment (no sampling)
         (default true)
  -log-json
        [boolean]
                Log in json format
         (default true)
  -log-level string
        [string]
                Log level: debug, info or error
    
  -proxy-edge-resource-store-namespace string
        [string]
                Resource store namespace to use
         (default "default")
  -proxy-edge-sync-interval duration
        [duration]
                Interval between synchronization cycles
         (default 1m0s)
  -proxy-grpc-router string
        [string]
                Which grpc proxy router to use
    
  -resource-store-client-certificate-path string
        [string]
                Path to PEM encoded certificate chain file
    
  -resource-store-client-enable-tls string
        [string]
                Enable mTLS from client-side
         (default "false")
  -resource-store-server-address string
        [string]
                gRPC server address to connect to
    
  -resource-store-server-name-override string
        [string]
                Server name used to verify the hostname returned by TLS handshake
    
  -server-dynamodb-resource-namespace-secondary-index string
        [string]
                DynamoDB ResourceStore namespace secondary index
         (default "resource_namespace")
  -server-dynamodb-resource-table-name string
        [string]
                DynamoDB ResourceStore table name to store resources
         (default "graviflow_resource_store")
  -server-envoy-xds-enable
        [boolean]
                Enable Envoy xDS server
    
  -server-envoy-xds-resource-store-namespace string
        [string]
                Resource store namespace to use
         (default "default")
  -server-envoy-xds-sync-interval duration
        [duration]
                Interval between synchronization cycles
         (default 1m0s)
  -server-resource-store-provider string
        [string]
                Resource store persistence layer provider
    
  -temporal-address string
        [string]
                Tempora server host:port
         (default "localhost:7233")
  -temporal-namespace string
        [string]
                Temporal namespace
         (default "default")