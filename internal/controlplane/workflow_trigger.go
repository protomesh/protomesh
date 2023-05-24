package controlplane

import (
	"context"
	"errors"
	"fmt"
	"strings"

	typesv1 "dev.azure.com/pomwm/pom-tech/graviflow/proto/api/types/v1"
	"github.com/google/uuid"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
	tmemporalsdk "go.temporal.io/sdk/temporal"
)

const (
	resourceNameAttr = "resourceName"
	resourceIdAttr   = "sourceName"
)

var (
	workflowIdNamespace = uuid.MustParse("3d8e41b4-f7d9-11ed-b67e-0242ac120002")
)

func (c *Controller[Dependency]) TriggerWorkflow(ctx context.Context, res *typesv1.Resource) error {

	rawSpec, err := res.Spec.UnmarshalNew()
	if err != nil {
		return err
	}

	spec, ok := rawSpec.(*typesv1.WorkflowTrigger)
	if !ok {
		return fmt.Errorf("Invalid resource type, expecting WorkflowTrigger, received: %s", res.Spec.TypeUrl)
	}

	logKvs := logKvsFromResource(res)
	log := c.Log()

	startOpts := client.StartWorkflowOptions{
		ID: spec.IdPrefix,
		SearchAttributes: map[string]interface{}{
			resourceIdAttr:   res.Id,
			resourceNameAttr: res.Name,
		},
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
	}

	// Build workflow ID
	switch idSuffix := spec.IdSuffix.(type) {

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
			uniqueId := uuid.NewSHA1(workflowIdNamespace, []byte(res.Id))
			startOpts.ID = strings.Join([]string{startOpts.ID, uniqueId.String()}, "")

		}

	default:
		return errors.New("Invalid ID suffix")

	}

	tempClient := c.Dependency().GetTemporalClient()

	// Check if workflow is running
	_, err = tempClient.DescribeWorkflowExecution(ctx, startOpts.ID, "")

	if _, ok := err.(*serviceerror.NotFound); err != nil && !ok {

		return err

	} else if err == nil {

		switch ifRunning := spec.IfRunning.(type) {

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
	if spec.ExecutionTimeout.IsValid() {
		if tout := spec.ExecutionTimeout.AsDuration(); tout > 0 {
			startOpts.WorkflowExecutionTimeout = tout
		}
	}

	if spec.RunTimeout.IsValid() {
		if tout := spec.RunTimeout.AsDuration(); tout > 0 {
			startOpts.WorkflowRunTimeout = tout
		}
	}

	if spec.TaskTimeout.IsValid() {
		if tout := spec.TaskTimeout.AsDuration(); tout > 0 {
			startOpts.WorkflowTaskTimeout = tout
		}
	}

	// Set retry policy
	if spec.RetryPolicy != nil {

		startOpts.RetryPolicy = &tmemporalsdk.RetryPolicy{}

		if spec.RetryPolicy.InitialInterval.IsValid() {
			if tout := spec.RetryPolicy.InitialInterval.AsDuration(); tout > 0 {
				startOpts.RetryPolicy.InitialInterval = tout
			}
		}

		if spec.RetryPolicy.MaximumBackoff.IsValid() {
			if tout := spec.RetryPolicy.MaximumBackoff.AsDuration(); tout > 0 {
				startOpts.RetryPolicy.MaximumInterval = tout
			}
		}

		if spec.RetryPolicy.MaximumAttempts > 0 {
			startOpts.RetryPolicy.MaximumAttempts = spec.RetryPolicy.MaximumAttempts
		}

		if spec.RetryPolicy.NonRetryableErrors != nil && len(spec.RetryPolicy.NonRetryableErrors) > 0 {
			startOpts.RetryPolicy.NonRetryableErrorTypes = spec.RetryPolicy.NonRetryableErrors
		}

	}

	// Build arguments slice
	args := []interface{}{}

	if argsList := spec.Arguments.GetListValue(); argsList != nil {
		args = argsList.AsSlice()
	}

	run, err := tempClient.ExecuteWorkflow(ctx, startOpts, spec.Name, args...)
	if err != nil {
		return err
	}

	log.Info("Workflow execution", append(logKvs, "workflowId", run.GetID(), "runId", run.GetRunID())...)

	return nil
}
