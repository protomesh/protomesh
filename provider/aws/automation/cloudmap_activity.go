package automation

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery/types"
)

type ListServicesFromCloudMapRequest struct {
	NamespaceName string            `json:"namespaceName,omitempty"`
	ServiceTags   map[string]string `json:"serviceTags,omitempty"`
	NextToken     string            `json:"nextToken,omitempty"`
}

type ListServicesFromCloudMapResponse struct {
	NextToken  string   `json:"nextToken,omitempty"`
	ServiceIds []string `json:"serviceIds,omitempty"`
}

func (as *AutomationSet[Dependency]) ListServicesFromCloudMap(ctx context.Context, req *ListServicesFromCloudMapRequest) (*ListServicesFromCloudMapResponse, error) {

	sdCli := as.Dependency().GetServiceDiscoveryClient()

	listNamespacesInput := &servicediscovery.ListNamespacesInput{
		Filters: []types.NamespaceFilter{
			types.NamespaceFilter{
				Name:      types.NamespaceFilterNameName,
				Condition: types.FilterConditionEq,
				Values:    []string{req.NamespaceName},
			},
		},
		MaxResults: aws.Int32(1),
	}

	listNamespaces, err := sdCli.ListNamespaces(ctx, listNamespacesInput)
	if err != nil {
		return nil, err
	}

	if len(listNamespaces.Namespaces) == 0 {

	}

	namespace := listNamespaces.Namespaces[0]

	listServicesInput := &servicediscovery.ListServicesInput{
		Filters: []types.ServiceFilter{
			types.ServiceFilter{
				Name:      types.ServiceFilterNameNamespaceId,
				Condition: types.FilterConditionEq,
				Values:    []string{aws.ToString(namespace.Name)},
			},
		},
		MaxResults: aws.Int32(25),
	}

	if len(req.NextToken) > 0 {
		listServicesInput.NextToken = aws.String(req.NextToken)
	}

	listServices, err := sdCli.ListServices(ctx, listServicesInput)
	if err != nil {
		return nil, err
	}

	res := &ListServicesFromCloudMapResponse{}

	if listServices.NextToken != nil {
		res.NextToken = aws.ToString(listNamespaces.NextToken)
	}

	for _, svc := range listServices.Services {

		if req.ServiceTags != nil {

			tagsForRes, err := sdCli.ListTagsForResource(ctx, &servicediscovery.ListTagsForResourceInput{
				ResourceARN: svc.Arn,
			})
			if err != nil {
				return nil, err
			}

			tags := make(map[string]string)

			for _, tag := range tagsForRes.Tags {

				key := aws.ToString(tag.Key)
				val := aws.ToString(tag.Value)

				if reqVal, ok := req.ServiceTags[key]; ok && reqVal == val {
					tags[key] = val
				}

			}

			if len(tags) != len(req.ServiceTags) {
				continue
			}

		}

	}

	return res, nil

}

type ListInstancesFromCloudMapServiceRequest struct {
	ServiceId string `json:"serviceId,omitempty"`
}

type CloudMapInstance struct {
	InstanceId string            `json:"instanceId,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

type ListInstancesFromCloudMapServiceResponse struct {
	Instances []*CloudMapInstance `json:"instances,omitempty"`
}

func (as *AutomationSet[Dependency]) ListInstancesFromCloudMapService(ctx context.Context, req *ListInstancesFromCloudMapServiceRequest) (*ListInstancesFromCloudMapServiceResponse, error) {

	sdCli := as.Dependency().GetServiceDiscoveryClient()

	listInstancesPages := servicediscovery.NewListInstancesPaginator(sdCli, &servicediscovery.ListInstancesInput{
		ServiceId: aws.String(req.ServiceId),
	})

	res := &ListInstancesFromCloudMapServiceResponse{
		Instances: make([]*CloudMapInstance, 0),
	}

	for listInstancesPages.HasMorePages() {

		listInstancePage, err := listInstancesPages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, instance := range listInstancePage.Instances {
			res.Instances = append(res.Instances, &CloudMapInstance{
				InstanceId: aws.ToString(instance.Id),
				Attributes: instance.Attributes,
			})
		}

	}

	return res, nil
}
