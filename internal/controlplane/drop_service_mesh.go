package controlplane

import (
	"context"

	typesv1 "dev.azure.com/pomwm/pom-tech/graviflow/proto/api/types/v1"
	apiv1 "dev.azure.com/pomwm/pom-tech/graviflow/proto/api/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

func (c *Controller[Dependency]) DropServiceMeshResourcesBeforeIndex(ctx context.Context, res *typesv1.Resource) error {

	logKvs := logKvsFromResource(res)
	log := c.Log()

	resCli := c.Dependency().GetResourceStoreClient()

	dropRes, err := resCli.DropBefore(ctx, &apiv1.DropBeforeResourcesRequest{
		Before:    res.Version.Index,
		Namespace: res.Namespace,
	})

	log.Info("Dropped resources before index", append(logKvs, "droppedCount", dropRes.DroppedCount)...)

	return err

}

func (c *Controller[Dependency]) DropServiceMeshResources(ctx context.Context, res *typesv1.Resource) error {

	list := new(structpb.ListValue)

	if err := res.Spec.UnmarshalTo(list); err != nil {
		return err
	}

	resourceIds := []string{}

	for _, item := range list.Values {

		resourceId := item.GetStringValue()

		if len(resourceId) > 0 {
			resourceIds = append(resourceIds, resourceId)
		}

	}

	logKvs := logKvsFromResource(res)
	log := c.Log()

	resCli := c.Dependency().GetResourceStoreClient()

	_, err := resCli.Drop(ctx, &apiv1.DropResourcesRequest{
		Namespace:   res.Namespace,
		ResourceIds: resourceIds,
	})

	log.Info("Dropped resources by id", logKvs)

	return err

}
