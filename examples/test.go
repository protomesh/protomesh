package main

import (
	"fmt"
	"time"

	typesv1 "github.com/upper-institute/graviflow/proto/api/types/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {

	typedRes := &typesv1.TypedResource{
		Namespace: "aaa",
		Spec: &typesv1.TypedResource_WorkflowTrigger{
			WorkflowTrigger: &typesv1.WorkflowTrigger{
				Name: "aaaaa",
				IfRunning: &typesv1.WorkflowTrigger_IfRunningAction_{
					IfRunningAction: typesv1.WorkflowTrigger_IF_RUNNING_ABORT,
				},
				IdSuffix: &typesv1.WorkflowTrigger_ExactIdSuffix{
					ExactIdSuffix: "awweqe",
				},
				TaskQueue: "graviflow",
				RetryPolicy: &typesv1.WorkflowTrigger_RetryPolicy{
					InitialInterval: durationpb.New(20 * time.Second),
					MaximumAttempts: 5,
				},
				Arguments: structpb.NewListValue(&structpb.ListValue{
					Values: []*structpb.Value{
						structpb.NewStructValue(&structpb.Struct{
							Fields: map[string]*structpb.Value{
								"zzz": structpb.NewStringValue("xxx"),
							},
						}),
					},
				}),
			},
		},
	}

	out, _ := protojson.Marshal(typedRes)

	fmt.Println(string(out))

	in := &typesaws.SyncS3ResourcesInput{
		ModifiedSince: timestamppb.Now(),
		Interval:      durationpb.New(65 * time.Second),
	}

	out, _ = protojson.Marshal(in)

	fmt.Println(string(out))

}
