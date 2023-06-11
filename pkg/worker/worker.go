package worker

import (
	"context"
	"errors"

	"github.com/protomesh/protomesh"
	"github.com/protomesh/protomesh/pkg/resource"
	servicesv1 "github.com/protomesh/protomesh/proto/api/services/v1"
	typesv1 "github.com/protomesh/protomesh/proto/api/types/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

type WorkerDependency interface {
	GetTemporalClient() client.Client
	GetResourceStoreClient() servicesv1.ResourceStoreClient
}

type Worker[Dependency WorkerDependency] struct {
	*protomesh.Injector[Dependency]

	WorkerTaskQueue protomesh.Config `config:"worker.task.queue,str" default:"protomesh" usage:"Temporal task queue to register activities and workflows"`

	SyncInterval           protomesh.Config `config:"sync.interval,duration" default:"60s" usage:"Interval between synchronization cycles"`
	ResourceStoreNamespace protomesh.Config `config:"resource.store.namespace,str" default:"default" usage:"Resource store namespace to use"`

	Worker worker.Worker

	//map[resourceId]
	triggers map[string]*typesv1.Trigger
}

func NewWorker[Dependency WorkerDependency]() *Worker[Dependency] {
	return &Worker[Dependency]{}
}

func (w *Worker[Dependency]) Initialize() {

	w.triggers = make(map[string]*typesv1.Trigger)

	workerOpts := worker.Options{
		WorkflowPanicPolicy:         worker.BlockWorkflow,
		DisableRegistrationAliasing: true,
	}

	w.Worker = worker.New(
		w.Dependency().GetTemporalClient(),
		w.WorkerTaskQueue.StringVal(),
		workerOpts,
	)
}

func (w *Worker[Depedency]) Start() {

	log := w.Log()

	err := w.Worker.Start()
	if err != nil {
		log.Panic("Temporal worker failed to start from controller (worker)", "error", err)
	}

}

func (w *Worker[Depedency]) Stop() {

	w.Worker.Stop()

}

func (w *Worker[Dependency]) BeforeBatch(ctx context.Context) error {
	return nil
}

func (w *Worker[Dependency]) OnUpdated(ctx context.Context, updatedRes *typesv1.Resource) error {

	pb, err := updatedRes.Spec.UnmarshalNew()
	if err != nil {
		return err
	}

	trigger, ok := pb.(*typesv1.Trigger)
	if !ok {
		return nil
	}

	w.triggers[updatedRes.Id] = trigger
	return nil

}

func (w *Worker[Dependency]) OnDropped(ctx context.Context, droppedRes *typesv1.Resource) error {

	trigger, ok := w.triggers[droppedRes.Id]
	if !ok {
		return nil
	}
	delete(w.triggers, droppedRes.Id)

	workflowId, err := workflowIdFromTrigger(trigger)
	if err != nil {
		return err
	}

	temporalCli := w.Dependency().GetTemporalClient()

	switch onDrop := trigger.OnDrop.(type) {

	case *typesv1.Trigger_OnDropAction_:

		var err error

		switch onDrop.OnDropAction {

		case typesv1.Trigger_ON_DROP_DO_NOTHING:
			return nil

		case typesv1.Trigger_ON_DROP_CANCEL:
			err = temporalCli.CancelWorkflow(ctx, workflowId, "")

		case typesv1.Trigger_ON_DROP_TERMINATE:
			err = temporalCli.TerminateWorkflow(ctx, workflowId, "", "Enforced by workflow trigger (Protomesh)")

		}

		switch {

		case errors.Is(err, &serviceerror.NotFound{}):
			return nil

		case err != nil:
			return err

		}

		return nil

	}

	return errors.New("Invalid on drop action")

}

func (w *Worker[Dependency]) AfterBatch(ctx context.Context) error {

	for resourceId, trigger := range w.triggers {

		log := w.Log().With("resourceId", resourceId)

		err := w.Trigger(ctx, trigger)
		if err != nil {
			log.Error("Trigger error", "error", err)
		}

	}

	return nil

}

func (w *Worker[Dependency]) Sync(ctx context.Context) <-chan error {

	sync := &resource.ResourceStoreSynchronizer[Dependency]{
		Injector:     w.Injector,
		SyncInterval: w.SyncInterval.DurationVal(),
		Namespace:    w.ResourceStoreNamespace.StringVal(),
		IndexCursor:  0,

		EventHandler: w,
	}

	return sync.Sync(ctx)

}
