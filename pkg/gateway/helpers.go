package gateway

import (
	typesv1 "github.com/protomesh/protomesh/proto/api/types/v1"
	"google.golang.org/protobuf/proto"
)

func getPolicySource(policy *typesv1.GatewayPolicy) (any, bool) {

	switch source := policy.Source.(type) {

	case *typesv1.GatewayPolicy_Grpc:
		return source.Grpc, source.Grpc.ExactMethodNameMatch

	case *typesv1.GatewayPolicy_Http:
		return source.Http, source.Http.ExactPathMatch

	default:
		panic("unreachable")

	}

}

func getHandlerFromOneOf(handler *typesv1.GatewayPolicy_Handler) (proto.Message, HandlerType) {

	switch handler := handler.Handler.(type) {

	case *typesv1.GatewayPolicy_Handler_Aws:

		switch handler := handler.Aws.Handler.(type) {

		case *typesv1.AwsHandler_Lambda:
			return handler.Lambda, HandlerTypeAwsLambda

		default:
			panic("unreachable")

		}

	default:
		panic("unreachable")

	}

}

func addPolicyToHandlerMap(handlerMap map[HandlerType][]*typesv1.GatewayPolicy, policy *typesv1.GatewayPolicy) {

	for _, handler := range policy.Handlers {
		_, handlerType := getHandlerFromOneOf(handler)

		if handlerType == HandlerTypeUndefined {
			continue
		}

		handlers, ok := handlerMap[handlerType]
		if !ok {
			handlers = make([]*typesv1.GatewayPolicy, 0)
		}

		handlerMap[handlerType] = append(handlers, policy)

	}

}

func fromHttpMethodToProtomeshHttpMethod(method string) typesv1.HttpMethod {

	switch method {

	case "GET":
		return typesv1.HttpMethod_HTTP_METHOD_GET

	case "POST":
		return typesv1.HttpMethod_HTTP_METHOD_POST

	case "PUT":
		return typesv1.HttpMethod_HTTP_METHOD_PUT

	case "PATCH":
		return typesv1.HttpMethod_HTTP_METHOD_PATCH

	case "DELETE":
		return typesv1.HttpMethod_HTTP_METHOD_DELETE

	case "HEAD":
		return typesv1.HttpMethod_HTTP_METHOD_HEAD

	case "OPTIONS":
		return typesv1.HttpMethod_HTTP_METHOD_OPTIONS

	default:
		return typesv1.HttpMethod_HTTP_METHOD_UNDEFINED

	}

}
