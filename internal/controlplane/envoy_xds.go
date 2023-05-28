package controlplane

import (
	"fmt"
	"sync"

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
	"github.com/upper-institute/graviflow"
	typesv1 "github.com/upper-institute/graviflow/proto/api/types/v1"
	"golang.org/x/net/context"
)

type EnvoyXdsDependency interface {
	GrpcServerProvider
	ResourceStoreProvider
}

type idNameMap struct {
	*cache.LinearCache

	idName map[string]string
	rw     *sync.RWMutex
}

func newIdNameMap(typeURL string) *idNameMap {
	return &idNameMap{
		LinearCache: cache.NewLinearCache(typeURL),
		idName:      make(map[string]string),
		rw:          &sync.RWMutex{},
	}
}

func (i *idNameMap) Set(key, val string) {
	i.rw.Lock()
	defer i.rw.Unlock()

	i.idName[key] = val
}

func (i *idNameMap) Get(key string) string {
	i.rw.RLock()
	i.rw.RUnlock()

	val, _ := i.idName[key]
	return val
}

func (i *idNameMap) Delete(key string) string {
	i.rw.Lock()
	defer i.rw.Unlock()

	val, ok := i.idName[key]

	if ok {
		delete(i.idName, key)
	}

	return val

}

type EnvoyXds[Dependency EnvoyXdsDependency] struct {
	*graviflow.AppInjector[Dependency]

	SyncInterval           graviflow.Config `config:"sync.interval,duration" default:"60s" usage:"Interval between synchronization cycles"`
	ResourceStoreNamespace graviflow.Config `config:"resource.store.namespace,str" default:"default" usage:"Resource store namespace to use"`

	resourceMap map[resource.Type]*idNameMap

	tx *envoyXdsTransaction
}

func (xds *EnvoyXds[Dependency]) Initialize() {

	xds.resourceMap = map[string]*idNameMap{
		resource.ListenerType:    newIdNameMap(resource.ListenerType),
		resource.ClusterType:     newIdNameMap(resource.ClusterType),
		resource.EndpointType:    newIdNameMap(resource.EndpointType),
		resource.VirtualHostType: newIdNameMap(resource.VirtualHostType),
		resource.RouteType:       newIdNameMap(resource.RouteType),
	}

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

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
		cache.Caches[typeUrl] = res
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
}

func (xds *EnvoyXds[Dependency]) OnBeforeProcess(ctx context.Context) error {

	xds.tx = &envoyXdsTransaction{
		operations: map[string]*envoyXdsTxOperation{},
	}

	for typeUrl := range xds.resourceMap {
		xds.tx.operations[typeUrl] = newEnvoyXdsTxOperation()
	}

	return nil

}

func (xds *EnvoyXds[Dependency]) OnUpdated(ctx context.Context, updatedRes *typesv1.Resource) error {

	res, err := updatedRes.Spec.UnmarshalNew()
	if err != nil {
		return err
	}

	name := cache.GetResourceName(res)
	if len(name) > 0 {

		typeUrl := updatedRes.Spec.TypeUrl

		op, ok := xds.tx.operations[typeUrl]
		if !ok {
			return fmt.Errorf("Unknown envoy xds type url: %s", typeUrl)
		}

		op.toUpdate[name] = res
		xds.resourceMap[typeUrl].Set(updatedRes.Id, name)

	}

	return nil

}

func (xds *EnvoyXds[Dependency]) OnDropped(ctx context.Context, droppedRes *typesv1.Resource) error {

	for typeUrl, res := range xds.resourceMap {

		if name := res.Delete(droppedRes.Id); len(name) > 0 {

			op := xds.tx.operations[typeUrl]

			op.toDelete = append(op.toDelete, name)

		}

	}

	return nil

}

func (xds *EnvoyXds[Dependency]) OnAfterProcess(ctx context.Context) error {

	for typeUrl, op := range xds.tx.operations {

		if len(op.toUpdate) > 0 || len(op.toDelete) > 0 {

			res := xds.resourceMap[typeUrl]

			if err := res.UpdateResources(op.toUpdate, op.toDelete); err != nil {
				return err
			}

		}

	}

	return nil

}

func (xds *EnvoyXds[Dependency]) Sync(ctx context.Context) <-chan error {

	sync := &ResourceStoreSynchronizer{
		SyncInterval:  xds.SyncInterval.DurationVal(),
		Namespace:     xds.ResourceStoreNamespace.StringVal(),
		ResourceStore: xds.Dependency().GetResourceStoreClient(),
		IndexCursor:   0,

		EventHandler: xds,
	}

	return sync.Sync(ctx)

}
