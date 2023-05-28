package controlplane

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/upper-institute/graviflow/internal/config"
	typesv1 "github.com/upper-institute/graviflow/proto/api/types/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
	tmemporalsdk "go.temporal.io/sdk/temporal"
)

const (
	workflowNameAttr = "workflowName"
)

var (
	workflowIdNamespace = uuid.MustParse("3d8e41b4-f7d9-11ed-b67e-0242ac120002")
)

func logKvsFromWorkflow(workTrigger *typesv1.WorkflowTrigger) []interface{} {
	return []interface{}{
		"workflowName", workTrigger.Name,
		"taskQueue", workTrigger.TaskQueue,
		"workflowIdPrefix", workTrigger.IdPrefix,
	}
}

func (c *Controller[Dependency]) TriggerWorkflowFromProtoJson(ctx context.Context, sourceFormat config.ProtoJson_SourceFormat, buf []byte) error {

	trigger := &typesv1.WorkflowTrigger{}

	err := config.ProtoJsonUnmarshal(buf, sourceFormat, trigger)
	if err != nil {
		return err
	}

	return c.TriggerWorkflow(ctx, trigger)
}

func (c *Controller[Dependency]) TriggerWorkflow(ctx context.Context, workTrigger *typesv1.WorkflowTrigger) error {

	logKvs := logKvsFromWorkflow(workTrigger)
	log := c.Log()

	startOpts := client.StartWorkflowOptions{
		ID: workTrigger.IdPrefix,
		SearchAttributes: map[string]interface{}{
			workflowNameAttr: workTrigger.Name,
		},
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
	}

	// Build workflow ID
	switch idSuffix := workTrigger.IdSuffix.(type) {

	case *typesv1.WorkflowTrigger_ExactIdSuffix:
		startOpts.ID = strings.Join([]string{startOpts.ID, idSuffix.ExactIdSuffix}, "")

	case *typesv1.WorkflowTrigger_IdSuffixBuilder:

		switch idSuffix.IdSuffixBuilder {

		case typesv1.WorkflowTrigger_ID_BUILDER_RANDOM:
			randomId, err := uuid.NewRandom()
			if err != nil {
				return err
			}
			startOpts.ID = strings.Join([]string{startOpts.ID, randomId.String()}, "")

		case typesv1.WorkflowTrigger_ID_BUILDER_UNIQUE:
			uniqueId := uuid.NewSHA1(workflowIdNamespace, []byte(workTrigger.IdPrefix))
			startOpts.ID = strings.Join([]string{startOpts.ID, uniqueId.String()}, "")

		}

	default:
		return errors.New("Invalid ID suffix")

	}

	tempClient := c.Dependency().GetTemporalClient()

	// Check if workflow is running
	_, err := tempClient.DescribeWorkflowExecution(ctx, startOpts.ID, "")

	if _, ok := err.(*serviceerror.NotFound); err != nil && !ok {

		return err

	} else if err == nil {

		switch ifRunning := workTrigger.IfRunning.(type) {

		case *typesv1.WorkflowTrigger_IfRunningAction_:

			switch ifRunning.IfRunningAction {

			case typesv1.WorkflowTrigger_IF_RUNNING_ABORT:
				log.Info("Workflow execution request aborted because there's an already running execution", logKvs...)
				return nil

			case typesv1.WorkflowTrigger_IF_RUNNING_OVERLAP:
				startOpts.WorkflowIDReusePolicy = enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING

			}

		default:
			return errors.New("Invalid if running specification")

		}

	}

	// Set timeouts
	if workTrigger.ExecutionTimeout.IsValid() {
		if tout := workTrigger.ExecutionTimeout.AsDuration(); tout > 0 {
			startOpts.WorkflowExecutionTimeout = tout
		}
	}

	if workTrigger.RunTimeout.IsValid() {
		if tout := workTrigger.RunTimeout.AsDuration(); tout > 0 {
			startOpts.WorkflowRunTimeout = tout
		}
	}

	if workTrigger.TaskTimeout.IsValid() {
		if tout := workTrigger.TaskTimeout.AsDuration(); tout > 0 {
			startOpts.WorkflowTaskTimeout = tout
		}
	}

	// Set retry policy
	if workTrigger.RetryPolicy != nil {

		startOpts.RetryPolicy = &tmemporalsdk.RetryPolicy{}

		if workTrigger.RetryPolicy.InitialInterval.IsValid() {
			if tout := workTrigger.RetryPolicy.InitialInterval.AsDuration(); tout > 0 {
				startOpts.RetryPolicy.InitialInterval = tout
			}
		}

		if workTrigger.RetryPolicy.MaximumBackoff.IsValid() {
			if tout := workTrigger.RetryPolicy.MaximumBackoff.AsDuration(); tout > 0 {
				startOpts.RetryPolicy.MaximumInterval = tout
			}
		}

		if workTrigger.RetryPolicy.MaximumAttempts > 0 {
			startOpts.RetryPolicy.MaximumAttempts = workTrigger.RetryPolicy.MaximumAttempts
		}

		if workTrigger.RetryPolicy.NonRetryableErrors != nil && len(workTrigger.RetryPolicy.NonRetryableErrors) > 0 {
			startOpts.RetryPolicy.NonRetryableErrorTypes = workTrigger.RetryPolicy.NonRetryableErrors
		}

	}

	// Build arguments slice
	args := []interface{}{}

	if argsList := workTrigger.Arguments.GetListValue(); argsList != nil {
		args = argsList.AsSlice()
	}

	run, err := tempClient.ExecuteWorkflow(ctx, startOpts, workTrigger.Name, args...)
	if err != nil {
		return err
	}

	log.Info("Workflow execution", append(logKvs, "workflowId", run.GetID(), "runId", run.GetRunID())...)

	return nil
}
