package aws

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"strconv"
	"time"

	"dev.azure.com/pomwm/pom-tech/graviflow"
	typesv1 "dev.azure.com/pomwm/pom-tech/graviflow/proto/api/types/v1"
	apiv1 "dev.azure.com/pomwm/pom-tech/graviflow/proto/api/v1"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type resourceStatus string

const (
	resourceActive  resourceStatus = "ACTIVE"
	resourceDropped resourceStatus = "DROPPED"
)

// Table with only partition key
type resourceDynamoDBRecord struct {
	ResourceId       string         `dynamodbav:"resourceId"`
	Name             string         `dynamodbav:"name"`
	VersionIndex     int64          `dynamodbav:"versionIndex"`
	Sha256Hash       string         `dynamodbav:"sha256Hash"`
	VersionTimestamp time.Time      `dynamodbav:"versionTimestamp,unixtime"`
	SpecTypeUrl      string         `dynamodbav:"specTypeUrl"`
	SpecValue        []byte         `dynamodbav:"specValue"`
	StoreNamespace   string         `dynamodbav:"storeNamespace"`
	ResourceStatus   resourceStatus `dynamodbav:"resourceStatus"`
}

func hashResource(res *typesv1.Resource) string {

	hash := sha256.New()

	raw, err := proto.Marshal(res)
	if err != nil {
		panic(err)
	}

	hash.Write(raw)

	return hex.EncodeToString(hash.Sum(nil))

}

type DynamoDBResourceStoreDependency interface {
	DynamoDBProvider
	GrpcServerProvider
}

type DynamoDBResourceStore[Dependency DynamoDBResourceStoreDependency] struct {
	graviflow.AppInjector[Dependency]
	apiv1.UnimplementedResourceStoreServer

	ResourceTableName               graviflow.Config `config:"resource.table.name,str" default:"graviflow_resource_store" usage:"DynamoDB ResourceStore table name to store resources"`
	ResourceNamespaceSecondaryIndex graviflow.Config `config:"resource.namespace.secondary.index,str" default:"resource_namespace" usage:"DynamoDB ResourceStore namespace secondary index"`
}

func (drs *DynamoDBResourceStore[Dependency]) Initialize() {

	log := drs.Log()

	dynamoCli := drs.Dependency().GetDynamoDBClient()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resourceTable := drs.ResourceTableName.StringVal()
	secondaryIndex := drs.ResourceNamespaceSecondaryIndex.StringVal()

	_, err := dynamoCli.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(resourceTable),
	})

	if err != nil {

		if errors.Is(err, &types.ResourceNotFoundException{}) {

			_, err = dynamoCli.CreateTable(ctx, &dynamodb.CreateTableInput{
				TableName:   aws.String(resourceTable),
				TableClass:  types.TableClassStandard,
				BillingMode: types.BillingModePayPerRequest,
				AttributeDefinitions: []types.AttributeDefinition{
					{AttributeName: aws.String("resourceId"), AttributeType: types.ScalarAttributeTypeS},
					{AttributeName: aws.String("storeNamespace"), AttributeType: types.ScalarAttributeTypeS},
					{AttributeName: aws.String("versionIndex"), AttributeType: types.ScalarAttributeTypeN},
				},
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("resourceId"), KeyType: types.KeyTypeHash},
				},
				GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
					{
						IndexName: aws.String(secondaryIndex),
						Projection: &types.Projection{
							ProjectionType: types.ProjectionTypeAll,
						},
						KeySchema: []types.KeySchemaElement{
							{
								AttributeName: aws.String("storeNamespace"),
								KeyType:       types.KeyTypeHash,
							},
							{
								AttributeName: aws.String("versionIndex"),
								KeyType:       types.KeyTypeRange,
							},
						},
					},
				},
				Tags: []types.Tag{{Key: aws.String("createdBy"), Value: aws.String("graviflow")}},
			})

			if err != nil {
				log.Panic("Error creating events table", "error", err)
			}

		}

		log.Panic("Error describing events table", "error", err)
	}

	apiv1.RegisterResourceStoreServer(drs.Dependency().GetGrpcServer(), drs)

}

func (drs *DynamoDBResourceStore[Dependency]) getResource(ctx context.Context, resourceId string, namespace string) (*resourceDynamoDBRecord, error) {

	dynamoCli := drs.Dependency().GetDynamoDBClient()

	proj := expression.NamesList(
		expression.Name("sha256Hash"),
		expression.Name("versionIndex"),
		expression.Name("versionTimestamp"),
		expression.Name("resourceStatus"),
		expression.Name("name"),
		expression.Name("specTypeUrl"),
		expression.Name("specValue"),
		expression.Name("storeNamespace"),
	)

	keyExp := expression.KeyEqual(expression.Key("resourceId"), expression.Value(resourceId))

	expr, err := expression.NewBuilder().
		WithKeyCondition(keyExp).
		WithProjection(proj).
		Build()
	if err != nil {
		return nil, err
	}

	getOut, err := dynamoCli.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(drs.ResourceTableName.StringVal()),
		Key: map[string]types.AttributeValue{
			"resourceId": &types.AttributeValueMemberS{
				Value: resourceId,
			},
		},
		ProjectionExpression:     expr.Projection(),
		ConsistentRead:           aws.Bool(false),
		ExpressionAttributeNames: expr.Names(),
	})
	if err != nil {
		return nil, err
	}

	if getOut.Item == nil || len(getOut.Item) == 0 {
		return nil, status.Error(codes.NotFound, "Resource not found by ID")
	}

	rec := &resourceDynamoDBRecord{}

	if err := attributevalue.UnmarshalMap(getOut.Item, rec); err != nil {
		return nil, err
	}

	if rec.StoreNamespace != namespace {
		return nil, status.Error(codes.NotFound, "Resource not found by namespace")
	}

	rec.ResourceId = resourceId

	return rec, nil

}

func (drs *DynamoDBResourceStore[Dependency]) Put(ctx context.Context, req *apiv1.PutResourceRequest) (*apiv1.PutResourceResponse, error) {

	rec, err := drs.getResource(ctx, req.Resource.Id, req.Resource.Namespace)

	dynamoCli := drs.Dependency().GetDynamoDBClient()

	resourceTable := drs.ResourceTableName.StringVal()

	sha256Hash := hashResource(&typesv1.Resource{
		Namespace: req.Resource.Namespace,
		Id:        req.Resource.Id,
		Name:      req.Resource.Name,
		Spec:      req.Resource.Spec,
	})

	versionTimestamp := time.Now()
	versionIndex := versionTimestamp.Unix()

	if err == nil {

		update := expression.
			Set(expression.Name("versionIndex"), expression.Value(versionIndex)).
			Set(expression.Name("versionTimestamp"), expression.Value(versionTimestamp))

		needUpdate := false

		if rec.Sha256Hash != sha256Hash {

			needUpdate = true

			update = update.
				Set(expression.Name("sha256Name"), expression.Value(sha256Hash)).
				Set(expression.Name("name"), expression.Value(req.Resource.Name)).
				Set(expression.Name("specTypeUrl"), expression.Value(req.Resource.Spec.TypeUrl)).
				Set(expression.Name("specValue"), expression.Value(req.Resource.Spec.Value))

		}

		if rec.ResourceStatus != resourceActive {

			needUpdate = true

			update = update.
				Set(expression.Name("resourceStatus"), expression.Value(resourceActive))

		}

		if needUpdate {

			expr, err := expression.NewBuilder().
				WithUpdate(update).
				Build()
			if err != nil {
				return nil, err
			}

			dynamoCli.UpdateItem(ctx, &dynamodb.UpdateItemInput{
				TableName:                 aws.String(resourceTable),
				ExpressionAttributeNames:  expr.Names(),
				ExpressionAttributeValues: expr.Values(),
				Key: map[string]types.AttributeValue{
					"resourceId": &types.AttributeValueMemberS{
						Value: req.Resource.Id,
					},
				},
				UpdateExpression: expr.Update(),
			})
		}

	} else if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {

		resourceRec, err := attributevalue.MarshalMap(&resourceDynamoDBRecord{
			ResourceId:       req.Resource.Id,
			VersionIndex:     versionIndex,
			Name:             req.Resource.Name,
			Sha256Hash:       sha256Hash,
			VersionTimestamp: versionTimestamp,
			SpecTypeUrl:      req.Resource.Spec.TypeUrl,
			SpecValue:        req.Resource.Spec.Value,
			StoreNamespace:   req.Resource.Namespace,
		})
		if err != nil {
			return nil, err
		}

		_, err = dynamoCli.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(resourceTable),
			Item:      resourceRec,
		})
		if err != nil {
			return nil, err
		}

	} else if err != nil {
		return nil, err
	}

	return &apiv1.PutResourceResponse{
		Version: &typesv1.Version{
			Sha256Hash: sha256Hash,
			Timestamp:  timestamppb.New(versionTimestamp),
			Index:      versionIndex,
		},
	}, nil

}

func (drs *DynamoDBResourceStore[Dependency]) dropResource(ctx context.Context, resourceId string, namespace string, versionTimestamp time.Time) error {

	dynamoCli := drs.Dependency().GetDynamoDBClient()

	resourceTable := drs.ResourceTableName.StringVal()

	versionIndex := versionTimestamp.Unix()

	update := expression.
		Set(expression.Name("resourceStatus"), expression.Value(resourceDropped)).
		Set(expression.Name("storeNamespace"), expression.Value(namespace)).
		Set(expression.Name("versionIndex"), expression.Value(versionIndex)).
		Set(expression.Name("versionTimestamp"), expression.Value(versionTimestamp))

	expr, err := expression.NewBuilder().
		WithUpdate(update).
		Build()
	if err != nil {
		return err
	}

	_, err = dynamoCli.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 aws.String(resourceTable),
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		Key: map[string]types.AttributeValue{
			"resourceId": &types.AttributeValueMemberS{
				Value: resourceId,
			},
		},
	})

	return err

}

func (drs *DynamoDBResourceStore[Dependency]) Drop(ctx context.Context, req *apiv1.DropResourcesRequest) (*apiv1.DropResourcesResponse, error) {

	versionTimestamp := time.Now()

	for _, resourceId := range req.ResourceIds {

		err := drs.dropResource(ctx, resourceId, req.Namespace, versionTimestamp)

		if err != nil {
			return nil, err
		}

	}

	return &apiv1.DropResourcesResponse{}, nil
}

func (drs *DynamoDBResourceStore[Dependency]) DropBefore(ctx context.Context, req *apiv1.DropBeforeResourcesRequest) (*apiv1.DropBeforeResourcesResponse, error) {

	dynamoCli := drs.Dependency().GetDynamoDBClient()

	resourceTable := drs.ResourceTableName.StringVal()
	secondaryIndex := drs.ResourceNamespaceSecondaryIndex.StringVal()

	pageSize := int32(100)

	proj := expression.NamesList(expression.Name("resourceId"))

	filter := expression.Equal(expression.Name("resourceStatus"), expression.Value(resourceActive))

	keyCond := expression.KeyEqual(expression.Key("storeNamespace"), expression.Value(req.Namespace))

	expr, err := expression.NewBuilder().
		WithProjection(proj).
		WithFilter(filter).
		WithKeyCondition(keyCond).
		Build()
	if err != nil {
		return nil, err
	}

	queryIn := &dynamodb.QueryInput{
		TableName:                 aws.String(resourceTable),
		ProjectionExpression:      expr.Projection(),
		KeyConditionExpression:    expr.Condition(),
		ConsistentRead:            aws.Bool(false),
		FilterExpression:          expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ExclusiveStartKey: map[string]types.AttributeValue{
			"storeNamespace": &types.AttributeValueMemberS{
				Value: req.Namespace,
			},
			"versionIndex": &types.AttributeValueMemberN{
				Value: strconv.FormatInt(req.Before, 10),
			},
		},
		ScanIndexForward: aws.Bool(false),
		Limit:            aws.Int32(pageSize),
		IndexName:        aws.String(secondaryIndex),
	}

	queryPages := dynamodb.NewQueryPaginator(dynamoCli, queryIn)

	droppedCount := int32(0)
	versionTimestamp := time.Now()

	for queryPages.HasMorePages() {

		queryPage, err := queryPages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, item := range queryPage.Items {

			resourceId := ""

			attributevalue.Unmarshal(item["resourceId"], &resourceId)

			if len(resourceId) == 0 {
				continue
			}

			droppedCount++

			if err := drs.dropResource(ctx, resourceId, req.Namespace, versionTimestamp); err != nil {
				return nil, err
			}

		}

	}

	return &apiv1.DropBeforeResourcesResponse{
		DroppedCount: droppedCount,
	}, nil

}

func (drs *DynamoDBResourceStore[Dependency]) List(stream apiv1.ResourceStore_ListServer) error {

	dynamoCli := drs.Dependency().GetDynamoDBClient()

	resourceTable := drs.ResourceTableName.StringVal()
	secondaryIndex := drs.ResourceNamespaceSecondaryIndex.StringVal()

	pageSize := int32(100)

	for {

		req, err := stream.Recv()
		if err != nil {

			if errors.Is(err, io.EOF) {
				break
			}

			return err

		}

		ctx := stream.Context()

		proj := expression.NamesList(
			expression.Name("sha256Hash"),
			expression.Name("versionIndex"),
			expression.Name("versionTimestamp"),
			expression.Name("resourceId"),
			expression.Name("resourceStatus"),
			expression.Name("name"),
			expression.Name("specTypeUrl"),
			expression.Name("specValue"),
		)

		keyCond := expression.KeyEqual(expression.Key("storeNamespace"), expression.Value(req.Namespace))

		expr, err := expression.NewBuilder().
			WithProjection(proj).
			WithKeyCondition(keyCond).
			Build()
		if err != nil {
			return err
		}

		queryIn := &dynamodb.QueryInput{
			TableName:                 aws.String(resourceTable),
			ProjectionExpression:      expr.Projection(),
			KeyConditionExpression:    expr.Condition(),
			ConsistentRead:            aws.Bool(false),
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
			ScanIndexForward:          aws.Bool(true),
			Limit:                     aws.Int32(pageSize),
			IndexName:                 aws.String(secondaryIndex),
		}

		if req.UpdatedSince > 0 {

			queryIn.ExclusiveStartKey = map[string]types.AttributeValue{
				"storeNamespace": &types.AttributeValueMemberS{
					Value: req.Namespace,
				},
				"versionIndex": &types.AttributeValueMemberN{
					Value: strconv.FormatInt(req.UpdatedSince, 10),
				},
			}

		}

		queryPages := dynamodb.NewQueryPaginator(dynamoCli, queryIn)

		for queryPages.HasMorePages() {

			queryPage, err := queryPages.NextPage(ctx)
			if err != nil {
				return err
			}

			res := &apiv1.ListResourcesResponse{
				UpdatedResources: []*typesv1.Resource{},
			}

			for _, item := range queryPage.Items {

				rec := &resourceDynamoDBRecord{}

				if err := attributevalue.UnmarshalMap(item, rec); err != nil {
					return err
				}

				switch rec.ResourceStatus {

				case resourceActive:
					res.UpdatedResources = append(res.UpdatedResources, &typesv1.Resource{
						Namespace: rec.StoreNamespace,
						Id:        rec.ResourceId,
						Name:      rec.Name,
						Spec: &anypb.Any{
							TypeUrl: rec.SpecTypeUrl,
							Value:   rec.SpecValue,
						},
						Version: &typesv1.Version{
							Sha256Hash: rec.Sha256Hash,
							Timestamp:  timestamppb.New(rec.VersionTimestamp),
							Index:      rec.VersionIndex,
						},
					})

				case resourceDropped:
					res.DroppedResources = append(res.DroppedResources, &typesv1.Resource{
						Namespace: rec.StoreNamespace,
						Id:        rec.ResourceId,
						Name:      rec.Name,
						Spec: &anypb.Any{
							TypeUrl: rec.SpecTypeUrl,
							Value:   rec.SpecValue,
						},
						Version: &typesv1.Version{
							Sha256Hash: rec.Sha256Hash,
							Timestamp:  timestamppb.New(rec.VersionTimestamp),
							Index:      rec.VersionIndex,
						},
					})

				}

			}

			if err := stream.Send(res); err != nil {
				return err
			}

		}

		if err := stream.Send(&apiv1.ListResourcesResponse{Nonce: req.Nonce}); err != nil {
			return err
		}

	}

	return nil
}
