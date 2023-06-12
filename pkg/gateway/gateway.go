package gateway

import (
	"context"

	"github.com/protomesh/protomesh"
	"github.com/protomesh/protomesh/pkg/resource"
	servicesv1 "github.com/protomesh/protomesh/proto/api/services/v1"
	typesv1 "github.com/protomesh/protomesh/proto/api/types/v1"
)

type GatewayHandler interface {
	// Context, active nodes and dropped nodes
	ProcessNodes(context.Context, []*typesv1.NetworkingNode, []*typesv1.NetworkingNode) error
}

type GatewayDependency interface {
	GetResourceStoreClient() servicesv1.ResourceStoreClient
}

type Gateway[D GatewayDependency] struct {
	*protomesh.Injector[D]

	ResourceStoreNamespace protomesh.Config `config:"resource.store.namespace,str" default:"default" usage:"Resource store namespace to use"`

	handlers []GatewayHandler

	updated []*typesv1.NetworkingNode
	dropped []*typesv1.NetworkingNode
}

func (ep *Gateway[D]) Initialize(handlers ...GatewayHandler) {

	ep.handlers = handlers

}

func (ep *Gateway[D]) BeforeBatch(ctx context.Context) error {

	ep.updated = []*typesv1.NetworkingNode{}
	ep.dropped = []*typesv1.NetworkingNode{}

	return nil

}

func (ep *Gateway[D]) OnUpdated(ctx context.Context, updatedRes *typesv1.Resource) error {

	edge := new(typesv1.NetworkingNode)

	err := updatedRes.Spec.UnmarshalTo(edge)
	if err != nil {
		return err
	}

	ep.updated = append(ep.updated, edge)

	return nil

}

func (ep *Gateway[D]) OnDropped(ctx context.Context, droppedRes *typesv1.Resource) error {

	edge := new(typesv1.NetworkingNode)

	err := droppedRes.Spec.UnmarshalTo(edge)
	if err != nil {
		return err
	}

	ep.dropped = append(ep.dropped, edge)

	return nil

}

func (ep *Gateway[D]) AfterBatch(ctx context.Context) error {

	for _, handler := range ep.handlers {

		if err := handler.ProcessNodes(ctx, ep.updated, ep.dropped); err != nil {
			return err
		}

	}

	return nil

}

func (ep *Gateway[D]) Sync(ctx context.Context) <-chan error {

	sync := &resource.ResourceStoreSynchronizer[D]{
		Injector:    ep.Injector,
		Namespace:   ep.ResourceStoreNamespace.StringVal(),
		IndexCursor: 0,

		EventHandler: ep,
	}

	return sync.Sync(ctx)

}
