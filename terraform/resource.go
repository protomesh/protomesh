package main

import (
	"context"

	"github.com/gogo/status"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	servicesv1 "github.com/protomesh/protomesh/proto/api/services/v1"
	typesv1 "github.com/protomesh/protomesh/proto/api/types/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func makeResourceSchema(nodeSchema map[string]*schema.Schema) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"namespace": {
			Description:  "Namespace in the resource store",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringLenBetween(1, 250),
			ForceNew:     true,
		},
		"resource_id": {
			Description:  "A valid UUID for the resource",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.IsUUID,
			ForceNew:     true,
		},
		"name": {
			Description:  "Name of the resource (display name)",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringLenBetween(1, 250),
		},
		"node": {
			Description: "Node specification.",
			Type:        schema.TypeList,
			MaxItems:    1,
			MinItems:    1,
			Required:    true,
			Elem: &schema.Resource{
				Schema: nodeSchema,
			},
		},
		"version_index": {
			Description: "Resource version index (unix timestamp in seconds)",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"version_sha256_sum": {
			Description: "Current version sha256 sum",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}

}

func putResource(ctx context.Context, rd *schema.ResourceData, i interface{}, res *typesv1.Resource) diag.Diagnostics {

	deps := i.(dependencies)

	resCli := deps.GetResourceStoreClient()

	putRes, err := resCli.Put(ctx, &servicesv1.PutResourceRequest{
		Resource: res,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	rd.SetId(rd.Get("resource_id").(string))

	rd.Set("version_index", putRes.Version.Index)
	rd.Set("version_sha256_sum", putRes.Version.Sha256Hash)

	return nil

}

func getResource(ctx context.Context, rd *schema.ResourceData, i interface{}) (*typesv1.Resource, diag.Diagnostics) {

	deps := i.(dependencies)

	resCli := deps.GetResourceStoreClient()

	rd.Get("namespace")

	getRes, err := resCli.Get(ctx, &servicesv1.GetResourceRequest{
		ResourceId: rd.Id(),
		Namespace:  rd.Get("namespace").(string),
	})
	if err != nil {

		if s, ok := status.FromError(err); ok && s.Code() == codes.NotFound {
			rd.SetId("")
			return nil, nil
		}

		return nil, diag.FromErr(err)
	}

	rd.Set("name", getRes.Resource.Name)
	rd.Set("version_index", getRes.Resource.Version.Index)
	rd.Set("version_sha256_sum", getRes.Resource.Version.Sha256Hash)

	return getRes.Resource, nil

}

func dropResource(ctx context.Context, rd *schema.ResourceData, i interface{}) diag.Diagnostics {

	deps := i.(dependencies)

	resCli := deps.GetResourceStoreClient()

	_, err := resCli.Drop(ctx, &servicesv1.DropResourcesRequest{
		ResourceIds: []string{rd.Id()},
		Namespace:   rd.Get("namespace").(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil

}

func makeDataSourceSchema(r *schema.Resource) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"namespace": {
			Description:  "Namespace in the resource store",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringLenBetween(1, 250),
			ForceNew:     true,
		},
		"resource_id": {
			Description:  "A valid UUID for the resource",
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.IsUUID,
			ForceNew:     true,
		},
		"name": {
			Description: "Name of the resource (display name)",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"spec": {
			Description: "Resource specification (payload)",
			Type:        schema.TypeList,
			Computed:    true,
			Elem:        r,
		},
		"version_index": {
			Description: "Resource version index (unix timestamp in seconds)",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"version_sha256_sum": {
			Description: "Current version sha256 sum",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}

}

func resourceFromResourceData(rd *schema.ResourceData, spec proto.Message) (*typesv1.Resource, diag.Diagnostics) {

	specAny, err := anypb.New(spec)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return &typesv1.Resource{
		Namespace: rd.Get("namespace").(string),
		Id:        rd.Get("resource_id").(string),
		Name:      rd.Get("name").(string),
		Spec:      specAny,
	}, nil

}

func nodeDataFromResourceData(rd *schema.ResourceData) (map[string]interface{}, diag.Diagnostics) {

	list, ok := rd.Get("node").([]interface{})
	if !ok || len(list) != 1 {
		return nil, diag.Errorf("Missing node definition")
	}

	return list[0].(map[string]interface{}), nil

}
