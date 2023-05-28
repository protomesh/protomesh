package main

import (
	"context"
	"io/ioutil"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery"
	"github.com/upper-institute/graviflow"
	"github.com/upper-institute/graviflow/internal/client"
	configpkg "github.com/upper-institute/graviflow/internal/config"
	"github.com/upper-institute/graviflow/internal/controlplane"
	apiv1 "github.com/upper-institute/graviflow/proto/api/v1"
	awsauto "github.com/upper-institute/graviflow/provider/aws/automation"
	temporalcli "go.temporal.io/sdk/client"
)

type ControllerDeps interface {
	GetTemporalClient() temporalcli.Client
	GetAwsConfig() aws.Config
}

type ControllerInjector interface {
	GetTemporalClient() temporalcli.Client
	GetResourceStoreClient() apiv1.ResourceStoreClient
	GetServiceDiscoveryClient() *servicediscovery.Client
	GetRoute53Client() *route53.Client
}

type ControllerInstance[D ControllerDeps] struct {
	*graviflow.AppInjector[D]

	ResourceStore       *client.GrpcClient[ControllerInjector] `config:"resource.store"`
	resourceStoreClient apiv1.ResourceStoreClient

	AwsAuto *awsauto.AutomationSet[ControllerInjector]

	Controls *controlplane.Controller[ControllerInjector] `config:"controls"`

	OnStartTriggerFile graviflow.Config `config:"on.start.file,str" usage:"The file path to proto json WorkflowTrigger file"`
}

func NewControllerInstance[D ControllerDeps]() *ControllerInstance[D] {
	return &ControllerInstance[D]{
		ResourceStore: &client.GrpcClient[ControllerInjector]{},
		AwsAuto:       &awsauto.AutomationSet[ControllerInjector]{},
		Controls:      &controlplane.Controller[ControllerInjector]{},
	}
}

func (c *ControllerInstance[D]) GetResourceStoreClient() apiv1.ResourceStoreClient {
	return c.resourceStoreClient
}

func (c *ControllerInstance[D]) GetServiceDiscoveryClient() *servicediscovery.Client {
	return servicediscovery.NewFromConfig(c.Dependency().GetAwsConfig())
}

func (c *ControllerInstance[D]) GetRoute53Client() *route53.Client {
	return route53.NewFromConfig(c.Dependency().GetAwsConfig())
}

func (c *ControllerInstance[D]) GetTemporalClient() temporalcli.Client {
	return c.Dependency().GetTemporalClient()
}

func (c *ControllerInstance[D]) Start() {

	c.ResourceStore.Start()

	c.resourceStoreClient = apiv1.NewResourceStoreClient(c.ResourceStore.ClientConn)

	c.Controls.Initialize()

	c.Controls.Register(
		c.AwsAuto,
	)

	c.Controls.Start()

	onStartFile := c.OnStartTriggerFile.StringVal()

	log := c.Log()

	if len(onStartFile) > 0 {

		buf, err := ioutil.ReadFile(onStartFile)
		if err != nil {
			log.Panic("Failed to read onStart trigger file", "error", err)
		}

		format, err := configpkg.ProtoJsonFileExtensionToFormat(onStartFile)
		if err != nil {
			log.Panic("Failed to parse extension of onStart trigger file", "error", err)
		}

		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		if err := c.Controls.TriggerWorkflowFromProtoJson(ctx, format, buf); err != nil {
			log.Panic("Failed to execute onStart trigger file", "error", err)
		}

	}

}

func (c *ControllerInstance[D]) Stop() {

	c.Controls.Stop()

	c.ResourceStore.Stop()

}
