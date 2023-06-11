package workflows

import (
	"time"

	"github.com/upper-institute/graviflow/internal/controlplane"
	typesaws "github.com/upper-institute/graviflow/proto/api/types/aws"
	typesv1 "github.com/upper-institute/graviflow/proto/api/types/v1"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func SyncS3Resources(ctx workflow.Context, input *typesaws.SyncS3ResourcesInput) error {

	scanReq := &typesaws.S3ScanResourcesRequest{
		BucketName:    input.BucketName,
		KeyPrefix:     input.KeyPrefix,
		ModifiedSince: input.ModifiedSince,
	}

	// now := workflow.Now(ctx)

	info := workflow.GetInfo(ctx)

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		TaskQueue:              info.TaskQueueName,
		ScheduleToStartTimeout: 30 * time.Second,
		StartToCloseTimeout:    info.WorkflowTaskTimeout,
		ScheduleToCloseTimeout: info.WorkflowTaskTimeout,
		HeartbeatTimeout:       0,
	})

	namespaceSet := map[string]bool{}

	for {

		scanCall := workflow.ExecuteActivity(ctx, "aws_S3ScanResources", scanReq)

		scanRes := &typesaws.S3ScanResourcesResponse{}

		if err := scanCall.Get(ctx, scanRes); err != nil {
			return err
		}

		for _, objectKey := range scanRes.ResourceObjects {

			objReq := &typesaws.S3ObjectToResourceRequest{
				BucketName: input.BucketName,
				ObjectKey:  objectKey,
			}

			objCall := workflow.ExecuteActivity(ctx, "aws_S3ObjectToResource", objReq)

			objRes := &typesaws.S3ObjectToResourceResponse{}

			if err := objCall.Get(ctx, objRes); err != nil {
				return err
			}

			typedRes := objRes.Resource

			rawRes := &typesv1.Resource{
				Namespace: typedRes.Namespace,
				Id:        typedRes.Id,
				Name:      typedRes.Name,
				Spec:      &anypb.Any{},
			}

			var controlOp workflow.Future

			switch spec := typedRes.Spec.(type) {

			case *typesv1.TypedResource_ServiceMeshEdge:

				if err := rawRes.Spec.MarshalFrom(spec.ServiceMeshEdge); err != nil {
					return err
				}

				controlOp = workflow.ExecuteActivity(ctx, "serviceMesh_PutEdge", rawRes)

			case *typesv1.TypedResource_ServiceMeshNode:

				if err := rawRes.Spec.MarshalFrom(spec.ServiceMeshNode); err != nil {
					return err
				}

				controlOp = workflow.ExecuteActivity(ctx, "serviceMesh_PutNode", rawRes)

			case *typesv1.TypedResource_WorkflowTrigger:

				cwo, err := controlplane.WorkflowTriggerToChildOptions(spec.WorkflowTrigger)
				if err != nil {
					return err
				}

				args := []interface{}{}

				if argsList := spec.WorkflowTrigger.Arguments.GetListValue(); argsList != nil {
					args = argsList.AsSlice()
				}

				child := workflow.ExecuteChildWorkflow(
					workflow.WithChildOptions(ctx, *cwo),
					spec.WorkflowTrigger.Name,
					args...,
				)

				// Wait for the Child Workflow Execution to spawn
				var childExec workflow.Execution
				if err := child.GetChildWorkflowExecution().Get(ctx, &childExec); err != nil {
					return err
				}

			}

			if controlOp == nil {
				continue
			}

			if err := controlOp.Get(ctx, nil); err != nil {
				return err
			}

			namespaceSet[typedRes.Namespace] = true

		}

		if len(scanRes.ContinuationToken) == 0 {
			break
		}

	}

	// for namespace := range namespaceSet {

	// 	if err := workflow.ExecuteActivity(ctx, "serviceMesh_DropResourcesBeforeIndex", &typesv1.Resource{
	// 		Namespace: namespace,
	// 		Version: &typesv1.Version{
	// 			Index: now.Unix(),
	// 		},
	// 	}).Get(ctx, nil); err != nil {
	// 		return err
	// 	}

	// }

	next := workflow.Now(ctx)

	if len(info.CronSchedule) > 0 {
		return nil
	}

	if err := workflow.Sleep(ctx, input.Interval.AsDuration()); err != nil {
		return err
	}

	return workflow.NewContinueAsNewError(ctx, "aws_SyncS3Resources", &typesaws.SyncS3ResourcesInput{
		BucketName:    input.BucketName,
		KeyPrefix:     input.KeyPrefix,
		ModifiedSince: timestamppb.New(next),
		Interval:      input.Interval,
	})

}
