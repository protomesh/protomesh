package main

import (
	"dev.azure.com/pomwm/pom-tech/graviflow"
	"dev.azure.com/pomwm/pom-tech/graviflow/internal/client"
	"dev.azure.com/pomwm/pom-tech/graviflow/internal/controlplane"
	apiv1 "dev.azure.com/pomwm/pom-tech/graviflow/proto/api/v1"
	awsauto "dev.azure.com/pomwm/pom-tech/graviflow/provider/aws/automation"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery"
	temporalcli "go.temporal.io/sdk/client"
)

type controllerDeps interface {
	GetTemporalClient() temporalcli.Client
	GetAwsConfig() aws.Config
}

type controllerInstance struct {
	graviflow.AppInjector[controllerDeps]

	resourceStore       *client.GrpcClient[*controllerInstance] `config:"resource.store"`
	resourceStoreClient apiv1.ResourceStoreClient

	awsAuto *awsauto.AutomationSet[*controllerInstance]

	controls *controlplane.Controller[*controllerInstance] `config:"controls"`
}

func newControllerInstance() *controllerInstance {
	return &controllerInstance{
		resourceStore: &client.GrpcClient[*controllerInstance]{},
		awsAuto:       &awsauto.AutomationSet[*controllerInstance]{},
		controls:      &controlplane.Controller[*controllerInstance]{},
	}
}

func (c *controllerInstance) GetResourceStoreClient() apiv1.ResourceStoreClient {
	return c.resourceStoreClient
}

func (c *controllerInstance) GetServiceDiscoveryClient() *servicediscovery.Client {
	return servicediscovery.NewFromConfig(c.Dependency().GetAwsConfig())
}

func (c *controllerInstance) GetRoute53Client() *route53.Client {
	return route53.NewFromConfig(c.Dependency().GetAwsConfig())
}

func (c *controllerInstance) GetTemporalClient() temporalcli.Client {
	return c.Dependency().GetTemporalClient()
}

func (c *controllerInstance) Start() {

	c.resourceStore.Start()

	c.resourceStoreClient = apiv1.NewResourceStoreClient(c.resourceStore.ClientConn)

	c.controls.Initialize()

	c.controls.Register(
		c.awsAuto,
	)

	c.controls.Start()

}

func (c *controllerInstance) Stop() {

	c.controls.Stop()

	c.resourceStore.Stop()

}
