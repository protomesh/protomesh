package automation

import (
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery"
	"github.com/upper-institute/graviflow"
	apiv1 "github.com/upper-institute/graviflow/proto/api/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"
)

type AutomationDependency interface {
	GetServiceDiscoveryClient() *servicediscovery.Client
	GetRoute53Client() *route53.Client
	GetResourceStoreClient() apiv1.ResourceStoreClient
}

type AutomationSet[Dependency AutomationDependency] struct {
	*graviflow.AppInjector[Dependency]
}

func (as *AutomationSet[Dependency]) Register(wk worker.Worker) {

	wk.RegisterActivityWithOptions(
		as.ListInstancesFromCloudMapService,
		activity.RegisterOptions{
			Name: "aws_listInstancesFromCloudMapService",
		},
	)

	wk.RegisterActivityWithOptions(
		as.ListServicesFromCloudMap,
		activity.RegisterOptions{
			Name: "aws_listServicesFromCloudMap",
		},
	)

	wk.RegisterActivityWithOptions(
		as.PutRoute53ZoneRecords,
		activity.RegisterOptions{
			Name: "aws_putRoute53ZoneRecord",
		},
	)

}
