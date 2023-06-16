package aws

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/protomesh/go-app"
	"github.com/protomesh/protomesh/pkg/server"
	typesv1 "github.com/protomesh/protomesh/proto/api/types/v1"
)

type GrpcLambdaRouterDependency interface {
	GetLambdaClient() *lambda.Client
}

type GrpcLambdaRouter[D GrpcLambdaRouterDependency] struct {
	*app.Injector[D]

	methods map[string]grpcLambdaMethodBuilder

	rwLock *sync.RWMutex
}

func (glr *GrpcLambdaRouter[D]) Initialize() {
	glr.methods = make(map[string]grpcLambdaMethodBuilder)
	glr.rwLock = new(sync.RWMutex)
}

// GetMethod returns a GrpcMethodHandler.
// It's the right place to implement things like A/B testing, traffic spliting and other routing strategies.
func (glr *GrpcLambdaRouter[D]) GetMethod(ctx context.Context, callInfo *server.GrpcCallInformation) server.GrpcMethodHandler {

	log := glr.Log()

	glr.rwLock.RLock()
	defer glr.rwLock.RUnlock()

	log.Debug("Call to gRPC method backed by a Lambda", "fullMethodName", callInfo.FullMethodName)

	methodBuilder, ok := glr.methods[callInfo.FullMethodName]
	if ok {
		return methodBuilder(ctx)
	}

	return nil

}

func (glr *GrpcLambdaRouter[D]) ProcessNodes(ctx context.Context, updated []*typesv1.NetworkingNode, dropped []*typesv1.NetworkingNode) error {

	for _, node := range updated {
		if err := glr.UpdateNode(ctx, node); err != nil {
			return err
		}
	}

	for _, node := range dropped {
		if err := glr.DropNode(ctx, node); err != nil {
			return err
		}
	}

	return nil

}

func (glr *GrpcLambdaRouter[D]) UpdateNode(ctx context.Context, node *typesv1.NetworkingNode) error {

	log := glr.Log()

	route, ok := node.NetworkingNode.(*typesv1.NetworkingNode_AwsLambdaGrpc)
	if !ok {
		return nil
	}

	glr.registerMethod(route.AwsLambdaGrpc)

	log.Debug(
		"New edge added to gRPC Lambda Router",
		"fullMethodName", route.AwsLambdaGrpc.FullMethodName,
		"functionName", route.AwsLambdaGrpc.FunctionName,
		"qualififer", route.AwsLambdaGrpc.Qualifier,
	)

	return nil

}

func (glr *GrpcLambdaRouter[D]) registerMethod(route *typesv1.AwsLambdaGrpc) {

	methodBuilder := newGrpcLambdaMethodBuilder(route, glr.Dependency().GetLambdaClient())

	glr.rwLock.Lock()
	defer glr.rwLock.Unlock()

	glr.methods[route.FullMethodName] = methodBuilder

}

func (glr *GrpcLambdaRouter[D]) DropNode(ctx context.Context, node *typesv1.NetworkingNode) error {

	log := glr.Log()

	route, ok := node.NetworkingNode.(*typesv1.NetworkingNode_AwsLambdaGrpc)
	if !ok {
		return nil
	}

	glr.unregisterMethod(route.AwsLambdaGrpc.FullMethodName)

	log.Debug(
		"Removed node from gRPC Lambda Router",
		"fullMethodName", route.AwsLambdaGrpc.FullMethodName,
	)

	return nil

}

func (glr *GrpcLambdaRouter[D]) unregisterMethod(fullMethodName string) {

	glr.rwLock.Lock()
	defer glr.rwLock.Unlock()

	delete(glr.methods, fullMethodName)

}
