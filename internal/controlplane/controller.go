package controlplane

import (
	"dev.azure.com/pomwm/pom-tech/graviflow"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"
)

type AutomationSet interface {
	Register(worker.Worker)
}

type ControllerDependency interface {
	TemporalProvider
	ResourceStoreProvider
}

type Controller[Dependency ControllerDependency] struct {
	graviflow.AppInjector[Dependency]

	WorkerTaskQueue graviflow.Config `config:"worker.task.queue,str" default:"graviflow" usage:"Temporal task queue to register activities and workflows"`

	worker.Worker
}

func NewController[Dependency ControllerDependency]() *Controller[Dependency] {
	return &Controller[Dependency]{}
}

func (c *Controller[Dependency]) Initialize() {

	workerOpts := worker.Options{
		WorkflowPanicPolicy:         worker.BlockWorkflow,
		DisableRegistrationAliasing: true,
	}

	c.Worker = worker.New(
		c.Dependency().GetTemporalClient(),
		c.WorkerTaskQueue.StringVal(),
		workerOpts,
	)
}

func (c *Controller[Dependency]) Register(sets ...AutomationSet) {

	for _, set := range sets {
		set.Register(c.Worker)
	}

}

func (c *Controller[Depedency]) Start() {

	log := c.Log()

	c.Worker.RegisterActivityWithOptions(
		c.PutServiceMeshNode,
		activity.RegisterOptions{
			Name: "serviceMesh_PutNode",
		},
	)

	c.Worker.RegisterActivityWithOptions(
		c.PutServiceMeshEdge,
		activity.RegisterOptions{
			Name: "serviceMesh_PutEdge",
		},
	)

	c.Worker.RegisterActivityWithOptions(
		c.DropServiceMeshResourcesBeforeIndex,
		activity.RegisterOptions{
			Name: "serviceMesh_DropResourcesBeforeIndex",
		},
	)

	c.Worker.RegisterActivityWithOptions(
		c.DropServiceMeshResources,
		activity.RegisterOptions{
			Name: "serviceMesh_DropResources",
		},
	)

	err := c.Worker.Start()
	if err != nil {
		log.Panic("Temporal worker failed to start from controller (controlplane)", "error", err)
	}

}

func (c *Controller[Depedency]) Stop() {

	c.Worker.Stop()

}
