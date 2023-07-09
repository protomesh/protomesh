package aws

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

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

	result       []byte
	resultCount  int
	isLastResult bool

	clientRequestPayload      []byte
	serverStreamTimeout       *time.Timer
	serverStreamTimeoutHeader string
	serverStreamError         error

	waitCall chan interface{}

	outgoingMetadata metadata.MD
}

func (l *lambdaGrpcHandler) Call(payload []byte) error {

	defer func() {
		l.waitCall <- nil
		if l.isLastResult {
			close(l.waitCall)
		}
	}()

	if l.serverStreamTimeout != nil {
		<-l.serverStreamTimeout.C
	}

	req := &events.APIGatewayProxyRequest{
		Path:              l.fullPath,
		MultiValueHeaders: l.incomingMetadata,
		Body:              base64.RawStdEncoding.EncodeToString(payload[:]),
		IsBase64Encoded:   true,
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

	if res.IsBase64Encoded {

		result, err := base64.RawStdEncoding.DecodeString(res.Body)
		if err != nil {
			return err
		}

		l.result = result

	} else {
		l.result = []byte(res.Body)
	}

	l.isLastResult = true

	switch res.StatusCode {

	// This status code is used for gRPC server-side streaming.
	case http.StatusProcessing:

		serverStreamTimeout, ok := l.outgoingMetadata[l.serverStreamTimeoutHeader]

		l.clientRequestPayload = payload
		l.isLastResult = false

		if ok && len(serverStreamTimeout) == 1 {

			d, err := time.ParseDuration(serverStreamTimeout[0])
			if err != nil {
				return status.Errorf(codes.Internal, "invalid %s header: %s", l.serverStreamTimeoutHeader, serverStreamTimeout)
			}

			if l.serverStreamTimeout == nil {
				l.serverStreamTimeout = time.NewTimer(d)
			} else {
				l.serverStreamTimeout.Reset(d)
			}

		}

		// Only the first call in the case of using a server-side stream
		// needs to return io.EOF, to end the client-side stream.
		// Subsequent calls will return nil, to continue the server-side stream.
		if l.resultCount > 0 {
			return nil
		}

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

	return io.EOF

}

func (l *lambdaGrpcHandler) Result() ([]byte, error) {
	<-l.waitCall

	l.resultCount++

	if l.isLastResult {
		return l.result, io.EOF
	}

	if l.serverStreamTimeout != nil {
		go func() {
			l.serverStreamError = l.Call(l.clientRequestPayload)
		}()
	}

	return l.result, l.serverStreamError
}

func (l *lambdaGrpcHandler) GetOutgoingMetadata() metadata.MD {
	return l.outgoingMetadata
}
