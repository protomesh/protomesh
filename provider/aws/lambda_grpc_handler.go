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
	"github.com/protomesh/protomesh/pkg/pubsub"
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

	subscriber pubsub.Subscriber[*LambdaNotification]
	pubsub     pubsub.PubSubSubscriber[*LambdaNotification]
}

func (l *lambdaGrpcHandler) attachPubSub(pubsubSubcriber pubsub.PubSubSubscriber[*LambdaNotification], subscriber pubsub.Subscriber[*LambdaNotification]) {

	l.subscriber = subscriber
	l.pubsub = pubsubSubcriber

}

func (l *lambdaGrpcHandler) unsubscribe() {

	if l.subscriber != nil {

		l.log.Debug("Unsubscribing from pubsub", "subscriber", l.subscriber.ID())

		l.pubsub.Unsubscribe(l.subscriber.ID())

	}

}

func (l *lambdaGrpcHandler) Call(payload []byte) error {

	log := l.log

	defer func() {
		l.waitCall <- nil
		if l.isLastResult {
			close(l.waitCall)
		}
	}()

	if l.serverStreamTimeout != nil {
		if l.subscriber == nil {
			<-l.serverStreamTimeout.C
		} else {
			select {
			case <-l.serverStreamTimeout.C:
			case <-l.subscriber.Stream():
			}
		}
	} else if l.subscriber != nil {
		<-l.subscriber.Stream()
	}

	req := &events.APIGatewayProxyRequest{
		Path:              l.fullPath,
		MultiValueHeaders: l.incomingMetadata,
		Body:              base64.RawStdEncoding.EncodeToString(payload[:]),
		IsBase64Encoded:   true,
	}

	in, err := json.Marshal(req)
	if err != nil {
		log.Error("Error marshalling request", "error", err)
		return err
	}

	out, err := l.lambdaCli.Invoke(l.ctx, &lambda.InvokeInput{
		FunctionName:   aws.String(l.param.FunctionName),
		InvocationType: types.InvocationTypeRequestResponse,
		Qualifier:      aws.String(l.param.Qualifier),
		Payload:        in,
	})
	if out != nil && out.FunctionError != nil {
		log.Error("Lambda function returned error", "error", aws.ToString(out.FunctionError))
		return fmt.Errorf("Lambda function returned error: %s", aws.ToString(out.FunctionError))
	}
	if err != nil {
		log.Error("Error invoking lambda function", "error", err)
		return err
	}

	res := &events.APIGatewayProxyResponse{}
	if err := json.Unmarshal(out.Payload, res); err != nil {
		log.Error("Error unmarshalling response", "error", err)
		return err
	}

	l.outgoingMetadata = metadata.Join(metadata.New(res.Headers), res.MultiValueHeaders)

	if res.IsBase64Encoded {

		result, err := base64.RawStdEncoding.DecodeString(res.Body)
		if err != nil {
			log.Error("Error decoding base64 response", "error", err)
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

		log.Debug("Lambda function returned server-side stream")

		serverStreamTimeout, ok := l.outgoingMetadata[l.serverStreamTimeoutHeader]

		l.clientRequestPayload = payload
		l.isLastResult = false

		if ok && len(serverStreamTimeout) == 1 {

			d, err := time.ParseDuration(serverStreamTimeout[0])
			if err != nil {
				log.Error("Error parsing server stream timeout", "error", err)
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
		log.Debug("Lambda function returned gone")
		return status.Error(codes.Aborted, res.Body)

	case http.StatusBadRequest:
		log.Debug("Lambda function returned bad request")
		return status.Error(codes.InvalidArgument, res.Body)

	case http.StatusPreconditionFailed:
		log.Debug("Lambda function returned precondition failed")
		return status.Error(codes.FailedPrecondition, res.Body)

	case http.StatusNotFound:
		log.Debug("Lambda function returned not found")
		return status.Error(codes.NotFound, res.Body)

	case http.StatusNotImplemented:
		log.Debug("Lambda function returned not implemented")
		return status.Error(codes.Unimplemented, res.Body)

	case http.StatusInternalServerError:
		log.Debug("Lambda function returned internal server error")
		return status.Error(codes.Internal, res.Body)

	case http.StatusGatewayTimeout:
		log.Debug("Lambda function returned gateway timeout")
		return status.Error(codes.DeadlineExceeded, res.Body)

	case http.StatusNoContent:
		log.Debug("Lambda function returned no content")
		return status.Error(codes.OutOfRange, res.Body)

	case http.StatusTooManyRequests:
		log.Debug("Lambda function returned too many requests")
		return status.Error(codes.ResourceExhausted, res.Body)

	case http.StatusServiceUnavailable:
		log.Debug("Lambda function returned service unavailable")
		return status.Error(codes.Unavailable, res.Body)

	case http.StatusConflict:
		log.Debug("Lambda function returned conflict")
		return status.Error(codes.AlreadyExists, res.Body)

	case http.StatusForbidden:
		log.Debug("Lambda function returned forbidden")
		return status.Error(codes.Unauthenticated, res.Body)

	case http.StatusInsufficientStorage:
		log.Debug("Lambda function returned insufficient storage")
		return status.Error(codes.DataLoss, res.Body)

	case http.StatusUnauthorized:
		log.Debug("Lambda function returned unauthorized")
		return status.Error(codes.PermissionDenied, res.Body)

	}

	log.Debug("Lambda function returned ok")

	return io.EOF

}

func (l *lambdaGrpcHandler) Result() ([]byte, error) {
	<-l.waitCall

	l.resultCount++

	if l.isLastResult {
		l.log.Debug("Lambda function returned last result")
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
	l.log.Debug("Lambda function returned outgoing metadata")
	return l.outgoingMetadata
}

func (l *lambdaGrpcHandler) Close() {
	l.unsubscribe()
}
