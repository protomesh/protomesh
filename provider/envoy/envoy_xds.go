package envoy

import (
	"fmt"

	clusterservice "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	endpointservice "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
	listenerservice "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	routeservice "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	runtimeservice "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	secretservice "github.com/envoyproxy/go-control-plane/envoy/service/secret/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/protomesh/go-app"
	resourcepkg "github.com/protomesh/protomesh/pkg/resource"
	servicesv1 "github.com/protomesh/protomesh/proto/api/services/v1"
	typesv1 "github.com/protomesh/protomesh/proto/api/types/v1"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type EnvoyXdsDependency interface {
	GetGrpcServer() *grpc.Server
	GetResourceStoreClient() servicesv1.ResourceStoreClient
}

type idNameMap struct {
	cache *cache.LinearCache

	idName map[string]string
}

func newIdNameMap(typeURL string, logger *cacheLogger) *idNameMap {
	return &idNameMap{
		cache:  cache.NewLinearCache(typeURL, cache.WithLogger(logger)),
		idName: make(map[string]string),
	}
}

func (i *idNameMap) Set(key, val string) {
	i.idName[key] = val
}

func (i *idNameMap) Get(key string) string {
	val, _ := i.idName[key]
	return val
}

func (i *idNameMap) Delete(key string) string {
	val, ok := i.idName[key]

	if ok {
		delete(i.idName, key)
	}

	return val

}

type cacheLogger struct {
	logger app.Logger
}

// Debugf logs a formatted debugging message.
func (c *cacheLogger) Debugf(format string, args ...interface{}) {
	c.logger.Debug(fmt.Sprintf(format, args...))
}

// Infof logs a formatted informational message.
func (c *cacheLogger) Infof(format string, args ...interface{}) {
	c.logger.Info(fmt.Sprintf(format, args...))
}

// Warnf logs a formatted warning message.
func (c *cacheLogger) Warnf(format string, args ...interface{}) {
	c.logger.Warn(fmt.Sprintf(format, args...))
}

// Errorf logs a formatted error message.
func (c *cacheLogger) Errorf(format string, args ...interface{}) {
	c.logger.Error(fmt.Sprintf(format, args...))
}

type EnvoyXds[Dependency EnvoyXdsDependency] struct {
	*app.Injector[Dependency]

	ResourceStoreNamespace app.Config `config:"resource.store.namespace,str" default:"default" usage:"Resource store namespace to use"`

	resourceMap map[resource.Type]*idNameMap

	instanceSetClusterNames *envoyClusters

	routing *routing

	tx *envoyXdsTransaction
}

func (xds *EnvoyXds[Dependency]) Initialize() {

	xds.instanceSetClusterNames = newEnvoyClusters()
	xds.routing = newRouting()

	caLogger := &cacheLogger{logger: xds.Log()}

	xds.resourceMap = map[string]*idNameMap{
		resource.ListenerType:    newIdNameMap(resource.ListenerType, caLogger),
		resource.ClusterType:     newIdNameMap(resource.ClusterType, caLogger),
		resource.EndpointType:    newIdNameMap(resource.EndpointType, caLogger),
		resource.VirtualHostType: newIdNameMap(resource.VirtualHostType, caLogger),
		resource.RouteType:       newIdNameMap(resource.RouteType, caLogger),
	}

	ctx := context.TODO()

	cache := &cache.MuxCache{
		Classify: func(r *cache.Request) string {
			return r.TypeUrl
		},
		ClassifyDelta: func(dr *cache.DeltaRequest) string {
			return dr.TypeUrl
		},
		Caches: map[string]cache.Cache{},
	}

	for typeUrl, res := range xds.resourceMap {
		cache.Caches[typeUrl] = res.cache
	}

	grpcServer := xds.Dependency().GetGrpcServer()

	xdsServer := server.NewServer(ctx, cache, nil)

	discoverygrpc.RegisterAggregatedDiscoveryServiceServer(grpcServer, xdsServer)

	endpointservice.RegisterEndpointDiscoveryServiceServer(grpcServer, xdsServer)
	clusterservice.RegisterClusterDiscoveryServiceServer(grpcServer, xdsServer)
	routeservice.RegisterRouteDiscoveryServiceServer(grpcServer, xdsServer)
	listenerservice.RegisterListenerDiscoveryServiceServer(grpcServer, xdsServer)
	secretservice.RegisterSecretDiscoveryServiceServer(grpcServer, xdsServer)
	runtimeservice.RegisterRuntimeDiscoveryServiceServer(grpcServer, xdsServer)
	routeservice.RegisterVirtualHostDiscoveryServiceServer(grpcServer, xdsServer)

}

type envoyXdsTxOperation struct {
	toUpdate map[string]types.Resource
	toDelete []string
}

func newEnvoyXdsTxOperation() *envoyXdsTxOperation {
	return &envoyXdsTxOperation{
		toUpdate: make(map[string]types.Resource),
		toDelete: make([]string, 0),
	}
}

type envoyXdsTransaction struct {
	operations map[resource.Type]*envoyXdsTxOperation
	endpoints  *clustersEndpoints
}

func (xds *EnvoyXds[Dependency]) BeforeBatch(ctx context.Context) error {

	xds.tx = &envoyXdsTransaction{
		operations: map[string]*envoyXdsTxOperation{},
		endpoints:  newClustersEndpoints(),
	}

	for typeUrl := range xds.resourceMap {
		xds.tx.operations[typeUrl] = newEnvoyXdsTxOperation()
	}

	return nil

}

func (xds *EnvoyXds[Dependency]) OnUpdated(ctx context.Context, updatedRes *typesv1.Resource) error {

	log := xds.Log()

	spec, err := updatedRes.Spec.UnmarshalNew()
	if err != nil {
		return err
	}

	switch spec := spec.(type) {

	case *typesv1.NetworkingNode:

		switch node := spec.NetworkingNode.(type) {

		case *typesv1.NetworkingNode_HttpIngress:

			list, routeConf, err := fromHttpIngress(node.HttpIngress)
			if err != nil {
				return err
			}

			name := cache.GetResourceName(list)

			xds.tx.operations[resource.ListenerType].toUpdate[name] = list
			xds.resourceMap[resource.ListenerType].Set(updatedRes.Id, name)

			xds.tx.operations[resource.RouteType].toUpdate[name] = routeConf
			xds.resourceMap[resource.RouteType].Set(updatedRes.Id, name)

		case *typesv1.NetworkingNode_InstanceSet:

			for _, clusterName := range node.InstanceSet.MatchServiceNames {

				xds.tx.endpoints.Add(clusterName, node.InstanceSet.Instances...)

				op := xds.tx.operations[resource.EndpointType]

				xds.instanceSetClusterNames.Add(clusterName, updatedRes.Id)

				id := createClusterLoadId(clusterName, updatedRes.Id)

				op.toUpdate[id] = xds.tx.endpoints.ToLoadAssignment(clusterName)

			}

			return nil

		case *typesv1.NetworkingNode_RoutingPolicy:

			xds.routing.putPolicy(updatedRes.Id, node.RoutingPolicy)

		case *typesv1.NetworkingNode_Service:
			svc, err := fromService(node.Service)
			if err != nil {
				return err
			}

			name := cache.GetResourceName(svc)

			xds.tx.operations[resource.ClusterType].toUpdate[name] = svc
			xds.resourceMap[resource.ClusterType].Set(updatedRes.Id, name)

		default:
			return nil

		}

	default:
		log.Warn("Invalid specification type for Envoy xDS server", "type_url", updatedRes.Spec.TypeUrl)

	}

	return nil

}

func (xds *EnvoyXds[Dependency]) OnDropped(ctx context.Context, droppedRes *typesv1.Resource) error {

	for typeUrl, res := range xds.resourceMap {

		op := xds.tx.operations[typeUrl]

		switch typeUrl {

		case resource.EndpointType:
			resMap, ok := xds.instanceSetClusterNames.Map[droppedRes.Id]
			if !ok {
				return nil
			}

			for clusterName := range resMap {

				id := createClusterLoadId(clusterName, droppedRes.Id)

				op.toDelete = append(op.toDelete, id)

			}

		case resource.VirtualHostType:
			xds.routing.dropPolicy(droppedRes.Id)

		default:
			if name := res.Delete(droppedRes.Id); len(name) > 0 {

				op.toDelete = append(op.toDelete, name)

			}

		}

	}

	return nil

}

func (xds *EnvoyXds[Dependency]) AfterBatch(ctx context.Context) error {

	xds.routing.processChanges()

	updatedVhs, droppedVhs := xds.routing.take()

	vhOp := xds.tx.operations[resource.VirtualHostType]

	vhOp.toUpdate = updatedVhs
	vhOp.toDelete = droppedVhs

	for typeUrl, op := range xds.tx.operations {

		if len(op.toUpdate) > 0 || len(op.toDelete) > 0 {

			res := xds.resourceMap[typeUrl]

			if err := res.cache.UpdateResources(op.toUpdate, op.toDelete); err != nil {
				return err
			}

		}

	}

	return nil

}

func (xds *EnvoyXds[Dependency]) OnEndOfPage(ctx context.Context) {

}

func (xds *EnvoyXds[Dependency]) Sync(ctx context.Context) <-chan error {

	sync := &resourcepkg.ResourceStoreSynchronizer[Dependency]{
		Injector:    xds.Injector,
		Namespace:   xds.ResourceStoreNamespace.StringVal(),
		IndexCursor: 0,

		EventHandler: xds,
	}

	return sync.Sync(ctx)

}
