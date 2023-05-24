package controlplane

import (
	"context"
	"fmt"

	typesv1 "dev.azure.com/pomwm/pom-tech/graviflow/proto/api/types/v1"
	apiv1 "dev.azure.com/pomwm/pom-tech/graviflow/proto/api/v1"
)

func (c *Controller[Dependency]) PutServiceMeshNode(ctx context.Context, res *typesv1.Resource) error {

	rawSpec, err := res.Spec.UnmarshalNew()
	if err != nil {
		return err
	}

	node, ok := rawSpec.(*typesv1.ServiceMesh_Node)
	if !ok {
		return fmt.Errorf("Invalid resource type, expecting ServiceMesh.PutNode, received: %s", res.Spec.TypeUrl)
	}

	resCli := c.Dependency().GetResourceStoreClient()

	resourceId := res.Id
	if len(resourceId) == 0 {
		resourceId = idFromServiceMeshNode(res.Namespace, node)
	}

	putReq := &apiv1.PutResourceRequest{
		Resource: &typesv1.Resource{
			Namespace: res.Namespace,
			Id:        resourceId,
			Name:      res.Name,
		},
	}

	switch node := node.Node.(type) {

	case *typesv1.ServiceMesh_Node_HttpIngress:

		putReq.Resource.Spec, err = fromHttpIngress(node)

	case *typesv1.ServiceMesh_Node_RoutingPolicy:

	case *typesv1.ServiceMesh_Node_Service:

		putReq.Resource.Spec, err = fromService(node)

	}

	if err != nil {
		return err
	}

	_, err = resCli.Put(ctx, &apiv1.PutResourceRequest{
		Resource: res,
	})
	return err

}

func (c *Controller[Dependency]) PutServiceMeshEdge(ctx context.Context, res *typesv1.Resource) error {

	rawSpec, err := res.Spec.UnmarshalNew()
	if err != nil {
		return err
	}

	_, ok := rawSpec.(*typesv1.ServiceMesh_Edge)
	if !ok {
		return fmt.Errorf("Invalid resource type, expecting ServiceMesh.Edge, received: %s", res.Spec.TypeUrl)
	}

	resCli := c.Dependency().GetResourceStoreClient()

	_, err = resCli.Put(ctx, &apiv1.PutResourceRequest{
		Resource: res,
	})

	return err

}
