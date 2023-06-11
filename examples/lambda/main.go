package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	servicesv1 "github.com/protomesh/protomesh/proto/api/services/v1"
	"google.golang.org/protobuf/proto"
)

func DoPong(ctx context.Context, req *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {

	msgIn := new(servicesv1.Ping)

	if err := proto.Unmarshal([]byte(req.Body), msgIn); err != nil {
		return nil, err
	}

	hash := sha256.New()

	hash.Write([]byte(msgIn.Nonce))

	msgOut := &servicesv1.Pong{
		Sha256Sum: base64.RawURLEncoding.EncodeToString(hash.Sum(nil)),
	}

	body, err := proto.Marshal(msgOut)
	if err != nil {
		return nil, err
	}

	res := &events.APIGatewayProxyResponse{
		Body: string(body),
	}

	return res, nil
}

func main() {
	lambda.Start(DoPong)
}
