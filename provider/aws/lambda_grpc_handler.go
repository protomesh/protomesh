package aws

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/protomesh/go-app"
	typesv1 "github.com/protomesh/protomesh/proto/api/types/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type lambdaGrpcHandler struct {
	log app.Logger

	fullPath string
	param    *typesv1.AwsHandler_LambdaFunction

	lambdaCli *lambda.Client

	ctx              context.Context
	incomingMetadata metadata.MD

	result           []byte
	waitCall         chan interface{}
	outgoingMetadata metadata.MD
}

func (l *lambdaGrpcHandler) Call(payload []byte) error {

	defer func() {
		l.waitCall <- nil
		close(l.waitCall)
	}()

	req := &events.APIGatewayProxyRequest{
		Path:              l.fullPath,
		MultiValueHeaders: l.incomingMetadata,
		Body:              string(payload[:]),
		IsBase64Encoded:   false,
	}

	in, err := json.Marshal(req)
	if err != nil {
		return err
	}

	out, err := l.lambdaCli.Invoke(l.ctx, &lambda.InvokeInput{
		FunctionName:   aws.String(l.param.FunctionName),
		InvocationType: types.InvocationTypeRequestResponse,
		Qualifier:      aws.String(l.param.Qualifier),
		Payload:        in,
	})
	if out != nil && out.FunctionError != nil {
		return fmt.Errorf("Lambda function returned error: %s", aws.ToString(out.FunctionError))
	}
	if err != nil {
		return err
	}

	res := &events.APIGatewayProxyResponse{}
	if err := json.Unmarshal(out.Payload, res); err != nil {
		return err
	}

	l.outgoingMetadata = metadata.Join(metadata.New(res.Headers), res.MultiValueHeaders)

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

		l.result = result

		return io.EOF

	}

	l.result = []byte(res.Body)

	return io.EOF

}

func (l *lambdaGrpcHandler) Result() ([]byte, error) {
	<-l.waitCall
	return l.result, io.EOF
}

func (l *lambdaGrpcHandler) GetOutgoingMetadata() metadata.MD {
	return l.outgoingMetadata
}
