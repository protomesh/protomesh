package workflows

import (
	"github.com/upper-institute/graviflow"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

type WorkflowsDependency interface {
	GetTemporalWorker() worker.Worker
}

type Workflows[D WorkflowsDependency] struct {
	*graviflow.AppInjector[D]
}

func (w *Workflows[D]) Register(wk worker.Worker) {

	wk.RegisterWorkflowWithOptions(
		SyncS3Resources,
		workflow.RegisterOptions{
			Name: "aws_SyncS3Resources",
		},
	)

}
