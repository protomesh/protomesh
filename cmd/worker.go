package main

import (
	protomesh "github.com/protomesh/protomesh"
	"github.com/protomesh/protomesh/pkg/client"
	workerpkg "github.com/protomesh/protomesh/pkg/worker"
	servicesv1 "github.com/protomesh/protomesh/proto/api/services/v1"
	temporalcli "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
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
	*protomesh.Injector[D]

	ResourceStore       *client.GrpcClient[WorkerInjector] `config:"resource.store"`
	resourceStoreClient servicesv1.ResourceStoreClient

	Worker *workerpkg.Worker[WorkerInjector] `config:"service"`
}

func NewWorkerInstance[D WorkerDeps]() WorkerInjector {
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

	w.ResourceStore.Start()

	w.resourceStoreClient = servicesv1.NewResourceStoreClient(w.ResourceStore.ClientConn)

	w.Worker.Initialize()

	w.Worker.Start()

}

func (c *WorkerInstance[D]) Stop() {

	c.Worker.Stop()

	c.ResourceStore.Stop()

}
