package main

import (
	"context"

	"github.com/protomesh/go-app"
	"github.com/protomesh/protomesh/pkg/client"
	workerpkg "github.com/protomesh/protomesh/pkg/worker"
	servicesv1 "github.com/protomesh/protomesh/proto/api/services/v1"
	temporalcli "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

var (
	_ WorkerInjector = &WorkerInstance[*root]{}
)

type WorkerDeps interface {
	GetTemporalClient() temporalcli.Client
}

type WorkerInjector interface {
	GetTemporalClient() temporalcli.Client
	GetTemporalWorker() worker.Worker
	GetResourceStoreClient() servicesv1.ResourceStoreClient

	Start()
	Stop()
}

type WorkerInstance[D WorkerDeps] struct {
	*app.Injector[D]

	ctx    context.Context
	cancel context.CancelFunc

	ResourceStore       *client.GrpcClient[WorkerInjector] `config:"resource.store"`
	resourceStoreClient servicesv1.ResourceStoreClient

	Worker      *workerpkg.Worker[WorkerInjector] `config:"service"`
	workerErrCh <-chan error
}

func NewWorkerInstance[D WorkerDeps]() *WorkerInstance[D] {
	return &WorkerInstance[D]{
		ResourceStore: &client.GrpcClient[WorkerInjector]{},
		Worker:        &workerpkg.Worker[WorkerInjector]{},
	}
}

func (w *WorkerInstance[D]) GetResourceStoreClient() servicesv1.ResourceStoreClient {
	return w.resourceStoreClient
}

func (w *WorkerInstance[D]) GetTemporalClient() temporalcli.Client {
	return w.Dependency().GetTemporalClient()
}

func (w *WorkerInstance[D]) GetTemporalWorker() worker.Worker {
	return w.Worker.Worker
}

func (w *WorkerInstance[D]) Start() {

	w.ctx, w.cancel = context.WithCancel(context.TODO())

	w.ResourceStore.Start()

	w.resourceStoreClient = servicesv1.NewResourceStoreClient(w.ResourceStore.ClientConn)

	w.Worker.Initialize()

	w.Worker.Start()

	w.workerErrCh = w.Worker.Sync(w.ctx)

}

func (w *WorkerInstance[D]) Stop() {

	w.cancel()

	<-w.workerErrCh

	w.Worker.Stop()

	w.ResourceStore.Stop()

}
