package worker

import (
	"context"
	"errors"

	"github.com/protomesh/protomesh/pkg/config"
	"github.com/protomesh/protomesh/pkg/logging"
	typesv1 "github.com/protomesh/protomesh/proto/api/types/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
	tmemporalsdk "go.temporal.io/sdk/temporal"
)

func (w *Worker[Dependency]) TriggerFromProtoJson(ctx context.Context, sourceFormat config.ProtoJson_SourceFormat, buf []byte) error {

	typedRes := &typesv1.TypedResource{}

	err := config.ProtoJsonUnmarshal(buf, sourceFormat, typedRes)
	if err != nil {
		return err
	}

	return w.Trigger(ctx, typedRes.GetTrigger())
}

func (w *Worker[Dependency]) Trigger(ctx context.Context, trigger *typesv1.Trigger) error {

	log := w.Log().With(logging.LogTrigger(trigger)...)

	startOpts, err := triggerToWorkflowStartOptions(trigger)
	if err != nil {
		return err
	}

	w.checkIfIsRunning(ctx, trigger, startOpts)

	tempClient := w.Dependency().GetTemporalClient()

	// Build arguments slice
	args := []interface{}{}

	if argsList := trigger.Arguments.GetListValue(); argsList != nil {
		args = argsList.AsSlice()
	}

	run, err := tempClient.ExecuteWorkflow(ctx, *startOpts, trigger.Name, args...)
	if err != nil {
		return err
	}

	log.Info("Workflow execution", "workflowId", run.GetID(), "runId", run.GetRunID())

	return nil

}

func (w *Worker[Dependency]) checkIfIsRunning(ctx context.Context, trigger *typesv1.Trigger, startOpts *client.StartWorkflowOptions) error {

	log := w.Log().With(logging.LogTrigger(trigger)...)

	tempClient := w.Dependency().GetTemporalClient()

	// Check if workflow is running
	runningWork, err := tempClient.DescribeWorkflowExecution(ctx, startOpts.ID, "")

	if _, ok := err.(*serviceerror.NotFound); err != nil && !ok {

		return err

	} else if err == nil {

		switch ifRunning := trigger.IfRunning.(type) {

		case *typesv1.Trigger_IfRunningAction_:

			switch ifRunning.IfRunningAction {

			case typesv1.Trigger_IF_RUNNING_ABORT:

				if runningWork.WorkflowExecutionInfo.Status == enums.WORKFLOW_EXECUTION_STATUS_RUNNING {
					log.Info("Workflow execution request aborted because there's an already running execution")
					return nil
				}
				startOpts.WorkflowIDReusePolicy = enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE

			case typesv1.Trigger_IF_RUNNING_OVERLAP:
				startOpts.WorkflowIDReusePolicy = enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING

			}

		default:
			return errors.New("Invalid if running specification")

		}

	}

	return nil

}

func triggerToWorkflowStartOptions(trigger *typesv1.Trigger) (*client.StartWorkflowOptions, error) {

	id, err := workflowIdFromTrigger(trigger)
	if err != nil {
		return nil, err
	}

	startOpts := &client.StartWorkflowOptions{
		ID:                    id,
		TaskQueue:             trigger.TaskQueue,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY,
	}

	if len(trigger.CronSchedule) > 0 {
		startOpts.CronSchedule = trigger.CronSchedule
	}

	// Set timeouts
	if trigger.ExecutionTimeout.IsValid() {
		if tout := trigger.ExecutionTimeout.AsDuration(); tout > 0 {
			startOpts.WorkflowExecutionTimeout = tout
		}
	}

	if trigger.RunTimeout.IsValid() {
		if tout := trigger.RunTimeout.AsDuration(); tout > 0 {
			startOpts.WorkflowRunTimeout = tout
		}
	}

	if trigger.TaskTimeout.IsValid() {
		if tout := trigger.TaskTimeout.AsDuration(); tout > 0 {
			startOpts.WorkflowTaskTimeout = tout
		}
	}

	// Set retry policy
	if trigger.RetryPolicy != nil {

		startOpts.RetryPolicy = &tmemporalsdk.RetryPolicy{}

		if trigger.RetryPolicy.InitialInterval.IsValid() {
			if tout := trigger.RetryPolicy.InitialInterval.AsDuration(); tout > 0 {
				startOpts.RetryPolicy.InitialInterval = tout
			}
		}

		if trigger.RetryPolicy.MaximumBackoff.IsValid() {
			if tout := trigger.RetryPolicy.MaximumBackoff.AsDuration(); tout > 0 {
				startOpts.RetryPolicy.MaximumInterval = tout
			}
		}

		if trigger.RetryPolicy.MaximumAttempts > 0 {
			startOpts.RetryPolicy.MaximumAttempts = trigger.RetryPolicy.MaximumAttempts
		}

		if trigger.RetryPolicy.NonRetryableErrors != nil && len(trigger.RetryPolicy.NonRetryableErrors) > 0 {
			startOpts.RetryPolicy.NonRetryableErrorTypes = trigger.RetryPolicy.NonRetryableErrors
		}

	}

	return startOpts, nil

}

// func triggerToChildOptions(workTrigger *typesv1.Trigger) (*workflow.ChildWorkflowOptions, error) {

// 	childOpts := &workflow.ChildWorkflowOptions{
// 		WorkflowID:            workTrigger.IdPrefix,
// 		TaskQueue:             workTrigger.TaskQueue,
// 		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
// 		ParentClosePolicy:     enums.PARENT_CLOSE_POLICY_ABANDON,
// 	}

// 	if len(workTrigger.CronSchedule) > 0 {
// 		childOpts.CronSchedule = workTrigger.CronSchedule
// 	}

// 	// Build workflow ID
// 	switch idSuffix := workTrigger.IdSuffix.(type) {

// 	case *typesv1.Trigger_ExactIdSuffix:
// 		childOpts.WorkflowID = strings.Join([]string{childOpts.WorkflowID, idSuffix.ExactIdSuffix}, "")

// 	case *typesv1.Trigger_IdSuffixBuilder:

// 		switch idSuffix.IdSuffixBuilder {

// 		case typesv1.Trigger_ID_BUILDER_RANDOM:
// 			randomId, err := uuid.NewRandom()
// 			if err != nil {
// 				return nil, err
// 			}
// 			childOpts.WorkflowID = strings.Join([]string{childOpts.WorkflowID, randomId.String()}, "")

// 		case typesv1.Trigger_ID_BUILDER_UNIQUE:
// 			uniqueId := uuid.NewSHA1(resource.WorkflowIdNamespace, []byte(workTrigger.IdPrefix))
// 			childOpts.WorkflowID = strings.Join([]string{childOpts.WorkflowID, uniqueId.String()}, "")

// 		}

// 	default:
// 		return nil, errors.New("Invalid ID suffix")

// 	}

// 	switch ifRunning := workTrigger.IfRunning.(type) {

// 	case *typesv1.Trigger_IfRunningAction_:

// 		switch ifRunning.IfRunningAction {

// 		case typesv1.Trigger_IF_RUNNING_ABORT:
// 			childOpts.WorkflowIDReusePolicy = enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY

// 		case typesv1.Trigger_IF_RUNNING_OVERLAP:
// 			childOpts.WorkflowIDReusePolicy = enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING

// 		}

// 	default:
// 		return nil, errors.New("Invalid if running specification")

// 	}

// 	// Set timeouts
// 	if workTrigger.ExecutionTimeout.IsValid() {
// 		if tout := workTrigger.ExecutionTimeout.AsDuration(); tout > 0 {
// 			childOpts.WorkflowExecutionTimeout = tout
// 		}
// 	}

// 	if workTrigger.RunTimeout.IsValid() {
// 		if tout := workTrigger.RunTimeout.AsDuration(); tout > 0 {
// 			childOpts.WorkflowRunTimeout = tout
// 		}
// 	}

// 	if workTrigger.TaskTimeout.IsValid() {
// 		if tout := workTrigger.TaskTimeout.AsDuration(); tout > 0 {
// 			childOpts.WorkflowTaskTimeout = tout
// 		}
// 	}

// 	// Set retry policy
// 	if workTrigger.RetryPolicy != nil {

// 		childOpts.RetryPolicy = &tmemporalsdk.RetryPolicy{}

// 		if workTrigger.RetryPolicy.InitialInterval.IsValid() {
// 			if tout := workTrigger.RetryPolicy.InitialInterval.AsDuration(); tout > 0 {
// 				childOpts.RetryPolicy.InitialInterval = tout
// 			}
// 		}

// 		if workTrigger.RetryPolicy.MaximumBackoff.IsValid() {
// 			if tout := workTrigger.RetryPolicy.MaximumBackoff.AsDuration(); tout > 0 {
// 				childOpts.RetryPolicy.MaximumInterval = tout
// 			}
// 		}

// 		if workTrigger.RetryPolicy.MaximumAttempts > 0 {
// 			childOpts.RetryPolicy.MaximumAttempts = workTrigger.RetryPolicy.MaximumAttempts
// 		}

// 		if workTrigger.RetryPolicy.NonRetryableErrors != nil && len(workTrigger.RetryPolicy.NonRetryableErrors) > 0 {
// 			childOpts.RetryPolicy.NonRetryableErrorTypes = workTrigger.RetryPolicy.NonRetryableErrors
// 		}

// 	}

// 	return childOpts, nil

// }
