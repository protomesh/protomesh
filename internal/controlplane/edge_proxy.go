package controlplane

import (
	"context"

	"github.com/upper-institute/graviflow"
	typesv1 "github.com/upper-institute/graviflow/proto/api/types/v1"
)

type EdgeProxyHandler interface {
	PutEdge(context.Context, *typesv1.ServiceMesh_Edge) error
	DropEdge(context.Context, *typesv1.ServiceMesh_Edge) error
}

type EdgeProxyDependency interface {
	ResourceStoreProvider
}

type EdgeProxy[Dependency EdgeProxyDependency] struct {
	*graviflow.AppInjector[Dependency]

	SyncInterval           graviflow.Config `config:"sync.interval,duration" default:"60s" usage:"Interval between synchronization cycles"`
	ResourceStoreNamespace graviflow.Config `config:"resource.store.namespace,str" default:"default" usage:"Resource store namespace to use"`

	handlers []EdgeProxyHandler

	updated []*typesv1.ServiceMesh_Edge
	dropped []*typesv1.ServiceMesh_Edge
}

func (ep *EdgeProxy[Dependency]) Initialize(handlers ...EdgeProxyHandler) {

	ep.handlers = handlers

}

func (ep *EdgeProxy[Dependency]) OnBeforeProcess(ctx context.Context) error {

	ep.updated = []*typesv1.ServiceMesh_Edge{}
	ep.dropped = []*typesv1.ServiceMesh_Edge{}

	return nil

}

func (ep *EdgeProxy[Dependency]) OnUpdated(ctx context.Context, updatedRes *typesv1.Resource) error {

	edge := new(typesv1.ServiceMesh_Edge)

	err := updatedRes.Spec.UnmarshalTo(edge)
	if err != nil {
		return err
	}

	ep.updated = append(ep.updated, edge)

	return nil

}

func (ep *EdgeProxy[Dependency]) OnDropped(ctx context.Context, droppedRes *typesv1.Resource) error {

	edge := new(typesv1.ServiceMesh_Edge)

	err := droppedRes.Spec.UnmarshalTo(edge)
	if err != nil {
		return err
	}

	ep.dropped = append(ep.dropped, edge)

	return nil

}

func (ep *EdgeProxy[Dependency]) OnAfterProcess(ctx context.Context) error {

	for _, handler := range ep.handlers {

		for _, updatedRes := range ep.updated {
			if err := handler.PutEdge(ctx, updatedRes); err != nil {
				return err
			}
		}

		for _, droppedRes := range ep.dropped {
			if err := handler.DropEdge(ctx, droppedRes); err != nil {
				return err
			}
		}

	}

	return nil

}

func (ep *EdgeProxy[Dependency]) Sync(ctx context.Context) <-chan error {

	sync := &ResourceStoreSynchronizer{
		SyncInterval:  ep.SyncInterval.DurationVal(),
		Namespace:     ep.ResourceStoreNamespace.StringVal(),
		ResourceStore: ep.Dependency().GetResourceStoreClient(),
		IndexCursor:   0,

		EventHandler: ep,
	}

	return sync.Sync(ctx)

}
