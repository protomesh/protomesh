package automation

import (
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery"
	"github.com/upper-institute/graviflow"
	apiv1 "github.com/upper-institute/graviflow/proto/api/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"
)

type AutomationDependency interface {
	GetAwsServiceDiscoveryClient() *servicediscovery.Client
	GetAwsRoute53Client() *route53.Client
	GetAwsS3Client() *s3.Client
	GetResourceStoreClient() apiv1.ResourceStoreClient
}

type AutomationSet[D AutomationDependency] struct {
	*graviflow.AppInjector[D]
}

func (as *AutomationSet[D]) Register(wk worker.Worker) {

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

	// S3
	wk.RegisterActivityWithOptions(
		as.S3ScanResources,
		activity.RegisterOptions{
			Name: "aws_S3ScanResources",
		},
	)

	wk.RegisterActivityWithOptions(
		as.S3ReadFile,
		activity.RegisterOptions{
			Name: "aws_S3ReadFile",
		},
	)

	wk.RegisterActivityWithOptions(
		as.S3ObjectToResource,
		activity.RegisterOptions{
			Name: "aws_S3ObjectToResource",
		},
	)

}
