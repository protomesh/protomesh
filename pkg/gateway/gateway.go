package gateway

import (
	"context"
	"net/http"

	"github.com/protomesh/go-app"
	"github.com/protomesh/protomesh/pkg/resource"
	servicesv1 "github.com/protomesh/protomesh/proto/api/services/v1"
	typesv1 "github.com/protomesh/protomesh/proto/api/types/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GatewayDependency interface {
	GetResourceStoreClient() servicesv1.ResourceStoreClient
}

type Gateway[D GatewayDependency] struct {
	*app.Injector[D]

	ResourceStoreNamespace app.Config `config:"resource.store.namespace,str" default:"default" usage:"Resource store namespace to use"`

	handlerMatcher handlerMatcher

	// Maps of resource id to policy.
	active  map[string]*typesv1.GatewayPolicy
	updated map[string]*typesv1.GatewayPolicy
	dropped map[string]*typesv1.GatewayPolicy

	grpcMatcher *sourceMatcher
	httpMatcher *sourceMatcher
}

func (g *Gateway[D]) Initialize(handlers ...GatewayHandler) {

	g.handlerMatcher = newHandlerMatcher()
	g.handlerMatcher.fromHandlersSlice(handlers)

	g.grpcMatcher = newSourceMatcher()
	g.httpMatcher = newSourceMatcher()

	g.active = make(map[string]*typesv1.GatewayPolicy)

}

func (g *Gateway[D]) BeforeBatch(ctx context.Context) error {

	g.updated = make(map[string]*typesv1.GatewayPolicy)
	g.dropped = make(map[string]*typesv1.GatewayPolicy)

	return nil

}

func (g *Gateway[D]) OnUpdated(ctx context.Context, updatedRes *typesv1.Resource) error {

	policy := new(typesv1.GatewayPolicy)

	err := updatedRes.Spec.UnmarshalTo(policy)
	if err != nil {
		g.Log().Warn("Skipping updated resource with invalid gateway spec", "resource", updatedRes.Id, "error", err)
		return nil
	}

	g.updated[updatedRes.Id] = policy

	return nil

}

func (g *Gateway[D]) OnDropped(ctx context.Context, droppedRes *typesv1.Resource) error {

	policy := new(typesv1.GatewayPolicy)

	err := droppedRes.Spec.UnmarshalTo(policy)
	if err != nil {
		g.Log().Warn("Skipping dropped resource with invalid gateway spec", "resource", droppedRes.Id, "error", err)
		return nil
	}

	g.dropped[droppedRes.Id] = policy

	return nil

}

func (g *Gateway[D]) AfterBatch(ctx context.Context) error {

	dropped := map[HandlerType][]*typesv1.GatewayPolicy{}
	updated := map[HandlerType][]*typesv1.GatewayPolicy{}

	for resourceId := range g.dropped {

		policy, ok := g.active[resourceId]
		if ok {

			delete(g.active, resourceId)

			switch policy.Source.(type) {

			case *typesv1.GatewayPolicy_Grpc:
				g.grpcMatcher.dropPolicy(policy)

			case *typesv1.GatewayPolicy_Http:
				g.httpMatcher.dropPolicy(policy)

			}

			addPolicyToHandlerMap(dropped, policy)

		}

	}

	for resourceId, policy := range g.updated {

		g.active[resourceId] = policy

		switch policy.Source.(type) {

		case *typesv1.GatewayPolicy_Grpc:
			g.grpcMatcher.addPolicy(policy)

		case *typesv1.GatewayPolicy_Http:
			g.httpMatcher.addPolicy(policy)

		}

		addPolicyToHandlerMap(updated, policy)

	}

	for handlerType, handler := range g.handlerMatcher {

		updatedPolicies, ok := updated[handlerType]
		if !ok {
			updatedPolicies = make([]*typesv1.GatewayPolicy, 0)
		}

		droppedPolicies, ok := dropped[handlerType]
		if !ok {
			updatedPolicies = make([]*typesv1.GatewayPolicy, 0)
		}

		if err := handler.ProcessPolicies(ctx, updatedPolicies, droppedPolicies); err != nil {
			return err
		}

	}

	return nil

}

func (g *Gateway[D]) Sync(ctx context.Context) <-chan error {

	sync := &resource.ResourceStoreSynchronizer[D]{
		Injector:    g.Injector,
		Namespace:   g.ResourceStoreNamespace.StringVal(),
		IndexCursor: 0,

		EventHandler: g,
	}

	return sync.Sync(ctx)

}

func (g *Gateway[D]) MatchGrpc(stream grpc.ServerStream) (*GrpcCall, error) {

	log := g.Log()

	fullMethodName, ok := grpc.MethodFromServerStream(stream)
	if !ok {
		return nil, status.Errorf(codes.Unimplemented, "Unknown method: %s", fullMethodName)
	}

	policy := g.grpcMatcher.matchPolicy(fullMethodName)
	if policy == nil {
		return nil, status.Errorf(codes.Unimplemented, "Unknown method: %s", fullMethodName)
	}

	call := &GrpcCall{
		Policy:   policy,
		Stream:   stream,
		Handlers: make([]GrpcHandler, 0),
	}

	for _, handlerOneOf := range policy.Handlers {

		handlerParam, handlerType := getHandlerFromOneOf(handlerOneOf)

		handler, ok := g.handlerMatcher[handlerType]
		if !ok {
			log.Error("Unknown handler type", "handlerType", handlerType)
			continue
		}

		call.Handlers = append(call.Handlers, handler.HandleGrpc(stream.Context(), handlerParam, call))

	}

	return call, nil

}

func (g *Gateway[D]) MatchHttp(res http.ResponseWriter, req *http.Request) (*HttpCall, error) {

	log := g.Log()

	methodName := typesv1.HttpMethod_name[int32(fromHttpMethodToProtomeshHttpMethod(req.Method))]

	policy := g.httpMatcher.matchPolicy(req.URL.Path, methodName)
	if policy == nil {
		return nil, http.ErrNotSupported
	}

	call := &HttpCall{
		Policy:   policy,
		Response: res,
		Request:  req,
		Handlers: make([]HttpHandler, 0),
	}

	for _, handlerOneOf := range policy.Handlers {

		handlerParam, handlerType := getHandlerFromOneOf(handlerOneOf)

		handler, ok := g.handlerMatcher[handlerType]
		if !ok {
			log.Error("Unknown handler type", "handlerType", handlerType)
			continue
		}

		call.Handlers = append(call.Handlers, handler.HandleHttp(req.Context(), handlerParam, call))

	}

	return call, nil

}
