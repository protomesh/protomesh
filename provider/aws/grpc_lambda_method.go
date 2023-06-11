package aws

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/protomesh/protomesh/pkg/server"
	typesv1 "github.com/protomesh/protomesh/proto/api/types/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type grpcLambdaMethodBuilder func(ctx context.Context) server.GrpcMethodHandler

type grpcLambdaMethod struct {
	route        *typesv1.AwsLambdaGrpc
	lambdaClient *lambda.Client

	ctx              context.Context
	incomingMetadata metadata.MD

	result           []byte
	waitCall         chan interface{}
	outgoingMetadata metadata.MD
}

func newGrpcLambdaMethodBuilder(route *typesv1.AwsLambdaGrpc, lambdaClient *lambda.Client) grpcLambdaMethodBuilder {

	return func(ctx context.Context) server.GrpcMethodHandler {

		incomingMetadata, _ := metadata.FromIncomingContext(ctx)

		return &grpcLambdaMethod{
			route:        route,
			lambdaClient: lambdaClient,

			ctx:              ctx,
			incomingMetadata: incomingMetadata,

			waitCall: make(chan interface{}),
		}
	}

}

func (g *grpcLambdaMethod) Call(payload []byte) error {

	defer func() {
		g.waitCall <- nil
		close(g.waitCall)
	}()

	req := &events.APIGatewayProxyRequest{
		Path:              g.route.FullMethodName,
		MultiValueHeaders: g.incomingMetadata,
		Body:              string(payload[:]),
		IsBase64Encoded:   false,
	}

	in, err := json.Marshal(req)
	if err != nil {
		return err
	}

	out, err := g.lambdaClient.Invoke(g.ctx, &lambda.InvokeInput{
		FunctionName:   aws.String(g.route.FunctionName),
		InvocationType: types.InvocationTypeRequestResponse,
		Qualifier:      aws.String(g.route.Qualifier),
		Payload:        in,
	})
	if err != nil {
		return err
	}

	res := &events.APIGatewayProxyResponse{}
	if err := json.Unmarshal(out.Payload, res); err != nil {
		return err
	}

	g.outgoingMetadata = metadata.Join(metadata.New(res.Headers), res.MultiValueHeaders)

	switch res.StatusCode {

	case http.StatusGone:
		return status.Error(codes.Aborted, res.Body)

	case http.StatusBadRequest:
		return status.Error(codes.InvalidArgument, res.Body)

	case http.StatusPreconditionFailed:
		return status.Error(codes.FailedPrecondition, res.Body)

	case http.StatusNotFound:
		return status.Error(codes.NotFound, res.Body)

	case http.StatusNotImplemented:
		return status.Error(codes.Unimplemented, res.Body)

	case http.StatusInternalServerError:
		return status.Error(codes.Internal, res.Body)

	case http.StatusGatewayTimeout:
		return status.Error(codes.DeadlineExceeded, res.Body)

	case http.StatusNoContent:
		return status.Error(codes.OutOfRange, res.Body)

	case http.StatusTooManyRequests:
		return status.Error(codes.ResourceExhausted, res.Body)

	case http.StatusServiceUnavailable:
		return status.Error(codes.Unavailable, res.Body)

	case http.StatusConflict:
		return status.Error(codes.AlreadyExists, res.Body)

	case http.StatusForbidden:
		return status.Error(codes.Unauthenticated, res.Body)

	case http.StatusInsufficientStorage:
		return status.Error(codes.DataLoss, res.Body)

	case http.StatusUnauthorized:
		return status.Error(codes.PermissionDenied, res.Body)

	}

	if res.IsBase64Encoded {

		result, err := base64.RawStdEncoding.DecodeString(res.Body)
		if err != nil {
			return err
		}

		g.result = result

	} else {
		g.result = []byte(res.Body)
	}

	return io.EOF

}

func (g *grpcLambdaMethod) GetResult() ([]byte, error) {

	<-g.waitCall

	return g.result, io.EOF

}

func (g *grpcLambdaMethod) GetOutgoingMetadata() metadata.MD {

	return g.outgoingMetadata

}
