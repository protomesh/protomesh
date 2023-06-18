package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/protomesh/go-app"
)

type AwsBuilder[D any] struct {
	*app.Injector[D]

	AwsConfig aws.Config

	DynamoDBEndpoint app.Config `config:"dynamodb.endpoint.url,str" usage:"Custom DynamoDB Endpoint url"`
}

func (a *AwsBuilder[D]) Initialize() {

	log := a.Log()

	dynamoDBEndpointUrl := a.DynamoDBEndpoint.StringVal()

	customResolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		if service == dynamodb.ServiceID {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           dynamoDBEndpointUrl,
				SigningRegion: "us-east-1",
			}, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	cfgOpts := []func(*config.LoadOptions) error{}

	if len(dynamoDBEndpointUrl) > 0 {

		log.Info("DynamoDB endpoint url", "endpointUrl", dynamoDBEndpointUrl)

		cfgOpts = append(cfgOpts, config.WithEndpointResolver(customResolver))

	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), cfgOpts...)
	if err != nil {
		log.Panic("Unable to load AWS configuration", "error", err)
	}

	a.AwsConfig = cfg

}
