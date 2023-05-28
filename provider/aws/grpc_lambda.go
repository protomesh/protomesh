package aws

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/upper-institute/graviflow"
	"github.com/upper-institute/graviflow/internal/server"
	typesv1 "github.com/upper-institute/graviflow/proto/api/types/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type grpcLambdaMethodHandler struct {
	*grpcLambdaMethod

	ctx     context.Context
	headers map[string]string
	result  []byte
}

func (gmh *grpcLambdaMethodHandler) Call(payload []byte) (metadata.MD, error) {

	req := &events.APIGatewayProxyRequest{
		Path:            gmh.fullMethodName,
		Headers:         gmh.headers,
		Body:            string(payload[:]),
		IsBase64Encoded: true,
	}

	in, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	out, err := gmh.lambdaClient.Invoke(gmh.ctx, &lambda.InvokeInput{
		FunctionName:   gmh.functionName,
		InvocationType: types.InvocationTypeRequestResponse,
		Qualifier:      gmh.qualifier,
		Payload:        in,
	})

	res := &events.APIGatewayProxyResponse{}
	if err := json.Unmarshal(out.Payload, res); err != nil {
		return nil, err
	}

	outMd := metadata.New(res.Headers)

	switch res.StatusCode {

	case http.StatusGone:
		return outMd, status.Error(codes.Aborted, res.Body)

	case http.StatusBadRequest:
		return outMd, status.Error(codes.InvalidArgument, res.Body)

	case http.StatusPreconditionFailed:
		return outMd, status.Error(codes.FailedPrecondition, res.Body)

	case http.StatusNotFound:
		return outMd, status.Error(codes.NotFound, res.Body)

	case http.StatusNotImplemented:
		return outMd, status.Error(codes.Unimplemented, res.Body)

	case http.StatusInternalServerError:
		return outMd, status.Error(codes.Internal, res.Body)

	case http.StatusGatewayTimeout:
		return outMd, status.Error(codes.DeadlineExceeded, res.Body)

	case http.StatusNoContent:
		return outMd, status.Error(codes.OutOfRange, res.Body)

	case http.StatusTooManyRequests:
		return outMd, status.Error(codes.ResourceExhausted, res.Body)

	case http.StatusServiceUnavailable:
		return outMd, status.Error(codes.Unavailable, res.Body)

	case http.StatusConflict:
		return outMd, status.Error(codes.AlreadyExists, res.Body)

	case http.StatusForbidden:
		return outMd, status.Error(codes.Unauthenticated, res.Body)

	case http.StatusInsufficientStorage:
		return outMd, status.Error(codes.DataLoss, res.Body)

	case http.StatusUnauthorized:
		return outMd, status.Error(codes.PermissionDenied, res.Body)

	}

	if res.IsBase64Encoded {

		result, err := base64.RawStdEncoding.DecodeString(res.Body)
		if err != nil {
			return nil, err
		}

		gmh.result = result

	} else {
		gmh.result = []byte(res.Body)
	}

	return outMd, io.EOF

}

func (gmh *grpcLambdaMethodHandler) Result() ([]byte, error) {

	return gmh.result, io.EOF

}

type grpcLambdaMethod struct {
	fullMethodName string
	functionName   *string
	qualifier      *string
	lambdaClient   *lambda.Client
}

func (glm *grpcLambdaMethod) Handle(ctx context.Context) (server.GrpcMethodHandler, error) {

	gmh := &grpcLambdaMethodHandler{
		ctx:              ctx,
		headers:          make(map[string]string),
		grpcLambdaMethod: glm,
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {

		for key, vals := range md {
			gmh.headers[key] = strings.Join(vals, ",")
		}

	}

	return gmh, nil

}

type GrpcLambdaRouterDependency interface {
	LambdaProvider
}

type GrpcLambdaRouter[Dependency GrpcLambdaRouterDependency] struct {
	graviflow.AppInjector[Dependency]

	methods map[string]*grpcLambdaMethod

	rwLock *sync.RWMutex
}

func (glr *GrpcLambdaRouter[Dependency]) Initialize() {
	glr.methods = make(map[string]*grpcLambdaMethod)
	glr.rwLock = new(sync.RWMutex)
}

func (glr *GrpcLambdaRouter[Dependency]) GetMethod(fullMethodName string) server.GrpcMethod {

	glr.rwLock.RLock()
	defer glr.rwLock.RUnlock()

	method, ok := glr.methods[fullMethodName]
	if ok {
		return method
	}

	return nil

}

func (glr *GrpcLambdaRouter[Dependency]) PutEdge(ctx context.Context, edge *typesv1.ServiceMesh_Edge) error {

	route, ok := edge.Edge.(*typesv1.ServiceMesh_Edge_AwsLambdaGrpcEdge)
	if !ok {
		return nil
	}

	glr.registerMethod(
		route.AwsLambdaGrpcEdge.FullMethodName,
		route.AwsLambdaGrpcEdge.FunctionName,
		route.AwsLambdaGrpcEdge.Qualifier,
	)

	return nil

}

func (glr *GrpcLambdaRouter[Dependency]) registerMethod(fullMethodName string, functionName string, qualifier string) {

	glr.rwLock.Lock()
	defer glr.rwLock.Unlock()

	method := &grpcLambdaMethod{
		fullMethodName: fullMethodName,
		functionName:   aws.String(functionName),
		lambdaClient:   glr.Dependency().GetLambdaClient(),
	}

	if len(qualifier) > 0 {
		method.qualifier = aws.String(qualifier)
	}

	glr.methods[fullMethodName] = method

}

func (glr *GrpcLambdaRouter[Dependency]) DropEdge(ctx context.Context, edge *typesv1.ServiceMesh_Edge) error {

	route, ok := edge.Edge.(*typesv1.ServiceMesh_Edge_AwsLambdaGrpcEdge)
	if !ok {
		return nil
	}

	glr.unregisterMethod(route.AwsLambdaGrpcEdge.FullMethodName)

	return nil

}

func (glr *GrpcLambdaRouter[Dependency]) unregisterMethod(fullMethodName string) {

	glr.rwLock.Lock()
	defer glr.rwLock.Unlock()

	delete(glr.methods, fullMethodName)

}
