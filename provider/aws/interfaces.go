package aws

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"google.golang.org/grpc"
)

type GrpcServerProvider interface {
	GetGrpcServer() *grpc.Server
}

type LambdaProvider interface {
	GetLambdaClient() *lambda.Client
}

type S3Provider interface {
	GetS3Client() *s3.Client
}

type DynamoDBProvider interface {
	GetDynamoDBClient() *dynamodb.Client
}
