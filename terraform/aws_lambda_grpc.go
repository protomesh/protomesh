package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	typesv1 "github.com/protomesh/protomesh/proto/api/types/v1"
	typesv1tf "github.com/protomesh/protomesh/terraform/proto/api/types/v1"
)

func dataSourceAwsLambdaGrpc() *schema.Resource {
	return &schema.Resource{
		ReadContext: getAwsLambdaGrpc,
	}
}

func resourceAwsLambdaGrpc() *schema.Resource {
	return &schema.Resource{
		Description:   "Expose a AWS Lambda through a gRPC interface",
		CreateContext: createAwsLambdaGrpc,
		ReadContext:   readAwsLambdaGrpc,
		UpdateContext: updateAwsLambdaGrpc,
		DeleteContext: deleteAwsLambdaGrpc,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: makeResourceSchema(typesv1tf.NewAwsLambdaGrpcSchema()),
	}
}

func createAwsLambdaGrpc(ctx context.Context, rd *schema.ResourceData, i interface{}) diag.Diagnostics {

	nodeData, diagErr := nodeDataFromResourceData(rd)
	if diagErr != nil {
		return diagErr
	}

	protoJsonMap, err := typesv1tf.UnmarshalAwsLambdaGrpc(nodeData)
	if err != nil {
		return diag.FromErr(err)
	}

	node := &typesv1.AwsLambdaGrpc{}

	if err := typesv1tf.UnmarshalAwsLambdaGrpcProto(protoJsonMap, node); err != nil {
		return diag.FromErr(err)
	}

	netNode := &typesv1.NetworkingNode{
		NetworkingNode: &typesv1.NetworkingNode_AwsLambdaGrpc{
			AwsLambdaGrpc: node,
		},
	}

	res, diagErr := resourceFromResourceData(rd, netNode)
	if diagErr != nil {
		return diagErr
	}

	return putResource(ctx, rd, i, res)

}

func updateAwsLambdaGrpc(ctx context.Context, rd *schema.ResourceData, i interface{}) diag.Diagnostics {

	return createAwsLambdaGrpc(ctx, rd, i)

}

func deleteAwsLambdaGrpc(ctx context.Context, rd *schema.ResourceData, i interface{}) diag.Diagnostics {

	return dropResource(ctx, rd, i)

}

func readAwsLambdaGrpc(ctx context.Context, rd *schema.ResourceData, i interface{}) diag.Diagnostics {

	res, diagErr := getResource(ctx, rd, i)
	if res == nil || diagErr != nil {
		return diagErr
	}

	netNode := &typesv1.NetworkingNode{}

	if err := res.Spec.UnmarshalTo(netNode); err != nil {
		return diag.FromErr(err)
	}

	if nodeSpec, ok := netNode.NetworkingNode.(*typesv1.NetworkingNode_AwsLambdaGrpc); ok {

		node, err := typesv1tf.MarshalAwsLambdaGrpcProto(nodeSpec.AwsLambdaGrpc)
		if err != nil {
			return diag.FromErr(err)
		}
		// panic(fmt.Sprintf("%+v", node))

		rd.Set("node", []interface{}{node})

	}

	return nil

}

func getAwsLambdaGrpc(ctx context.Context, rd *schema.ResourceData, i interface{}) diag.Diagnostics {

	rd.SetId(rd.Get("resource_id").(string))

	return readAwsLambdaGrpc(ctx, rd, i)

}
