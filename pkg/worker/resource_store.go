package worker

import (
	"context"

	servicesv1 "github.com/protomesh/protomesh/proto/api/services/v1"
)

func (w *Worker[Dependency]) Put(ctx context.Context, in *servicesv1.PutResourceRequest) (*servicesv1.PutResourceResponse, error) {
	return w.Dependency().GetResourceStoreClient().Put(ctx, in)
}

func (w *Worker[Dependency]) Drop(ctx context.Context, in *servicesv1.DropResourcesRequest) (*servicesv1.DropResourcesResponse, error) {
	return w.Dependency().GetResourceStoreClient().Drop(ctx, in)
}

func (w *Worker[Dependency]) Get(ctx context.Context, in *servicesv1.GetResourceRequest) (*servicesv1.GetResourceResponse, error) {
	return w.Dependency().GetResourceStoreClient().Get(ctx, in)
}
