# Protomesh

Protomesh is a Cloud Automation tool focused on managing environments for cloud native applications.

- [Protomesh](#protomesh)
  - [Components](#components)
  - [Build](#build)
  - [Configuration](#configuration)
  - [Deploy](#deploy)
    - [Deploy Resource Store](#deploy-resource-store)
    - [Deploy Envoy xDS](#deploy-envoy-xds)
    - [Deploy Gateway](#deploy-gateway)
    - [Deploy Worker](#deploy-worker)
  - [Resources](#resources)
  - [Building example AWS Lambda](#building-example-aws-lambda)

## Components

There are 4 software components of Protomesh:

- **Protomesh Resource Store:** a gRPC service to manage and provide resource state to other components.
- **Protomesh Envoy xDS:** a gRPC [Envoy discovery service](https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol) to configure Envoy instances in the service mesh.
- **Protomesh Gateway:** a high-level application gateway to enable features as such calling AWS Lambda from gRPC methods.
- **Protomesh Worker:** a [Temporal](https://temporal.io/) worker to ensure that automation workflows are running as configured in the resource store.

## Build

To build the application we recommend reading the [Containerfile](./Containerfile) which is used to generate automated releases of container images in this repository. But here is a short guide to setup your local environment and build Protomesh. All four components are provided by a single binary.

1. Download all modules for this repo: `go mod download`
2. Build the application: `go build -o ./cmd/protomesh ./cmd`

And then in the **cmd/** repository you'll find the executable **protomesh**. In this directory you can find configurations for each component in form of TOML files. These files are used by the [docker-compose.yaml](./docker-compose.yaml) definition file.

## Configuration

You can run `./protomesh -h` to get a list of all available flags. Each flag can be provided in the form of environment variables, you just need to convert the flag to upper snake case, example: a flag named `-my-parameter` will be read from environment variable `MY_PARAMETER`.

The precedence of configuration follows the order specified bellow (from most relevant to less relevant):

1. Command line flags, example: `-my-flag=true`
2. Environment variable, example: `MY_FLAG=true`
3. Configuration file (YAML, JSON or TOML), example: `{"my":{"flag":true}}`

## Deploy

We recommend running Protomesh components using Containers. Protomesh does not require so much of CPU and memory resources, but is designed to handle heavy networking IO volumes and concurrency.

There is only one required component in Protomesh deployment: Resource Store. This component provides a gRPC interface (default in the 6680 TCP port) for the [Resource Store](./api/services/v1/resource_store.proto) service. It's just an abstraction layer for resource persistence in technologies such as SQL databases, object stores. For now, we only support Postgres, but it's easy to implement a new persistence layer, you just need to develop a gRPC service for Resource Store using the Protobuf definition in this repository.

All other components connects to Protomesh Resource Store to synchronize its internal states.

By default we provide recommendations for the following deployment scenarios:

1. **Docker Compose:** ideal to run Protomesh components locally.
2. **Cloud Services:** using CaaS like AWS ECS.
3. **Kubernetes:** for now, there's no documented o pre-built assets to deploy Protomesh in Kubernetes, if you want to contribute please, open an issue.

### Deploy Resource Store

### Deploy Envoy xDS

### Deploy Gateway

### Deploy Worker

## Resources

All resources are documented in the [Terraform provider](https://registry.terraform.io/providers/protomesh/protomesh/latest/docs).

## Building example AWS Lambda

To build the example AWS Lambda that handles the `PingPong.DoPing` gRPC call use the following commands.

```bash
cd examples/lambda
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o main .
rm -rf main.zip
zip main.zip main
```