package aws

import (
	"context"

	"dev.azure.com/pomwm/pom-tech/graviflow"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

type AwsBuilder[Dependency any] struct {
	graviflow.AppInjector[Dependency]

	aws.Config

	EnableGrpcLambdaRouter graviflow.Config `config:"enable.grpc.lambda.router,bool" default:"false" usage:"Enable gRPC Lambda router"`
}

func (a *AwsBuilder[Dependency]) Initialize() {

	log := a.Log()

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Panic("Unable to load AWS configuration", "error", err)
	}

	a.Config = cfg

}
