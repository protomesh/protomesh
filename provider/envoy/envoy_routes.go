package envoy

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"strings"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/iancoleman/strcase"
	typesv1 "github.com/protomesh/protomesh/proto/api/types/v1"
	"google.golang.org/protobuf/proto"
)

func hashRoute(route *typesv1.RoutingPolicy_Route, suffixes ...string) string {

	h := sha256.New()

	pb, _ := proto.Marshal(route)

	h.Write(pb)

	for _, suffix := range suffixes {
		io.WriteString(h, suffix)
	}

	return hex.EncodeToString(h.Sum(nil))

}

func toEnvoyRoute(route *typesv1.RoutingPolicy_Route) *routev3.Route {
	return &routev3.Route{
		Match: &routev3.RouteMatch{
			PathSpecifier: &routev3.RouteMatch_Prefix{
				Prefix: route.MatchPrefix,
			},
		},
		Action: &routev3.Route_Route{
			Route: &routev3.RouteAction{
				Timeout: route.Timeout,
				ClusterSpecifier: &routev3.RouteAction_Cluster{
					Cluster: route.TargetService,
				},
				PrefixRewrite: route.PrefixRewrite,
			},
		},
	}
}

type virtualHostExt struct {
	*routev3.VirtualHost

	routeMap map[string]interface{}

	sha256sum []byte
}

func (v *virtualHostExt) hash() {

	h := sha256.New()

	pb, _ := proto.Marshal(v.VirtualHost)

	h.Write(pb)

	v.sha256sum = h.Sum(nil)

}

func (v *virtualHostExt) isEqual(b *virtualHostExt) bool {
	return bytes.Equal(b.sha256sum, v.sha256sum)
}

type routing struct {
	// map[virtualHostName]
	updated map[string]*routev3.VirtualHost

	// map[virtualHostName]
	dropped map[string]interface{}

	// map[resourceId]
	resources map[string]*typesv1.RoutingPolicy

	// map[virtualHostName]
	virtualHosts map[string]*virtualHostExt
}

func newRouting() *routing {
	return &routing{
		updated:      make(map[string]*routev3.VirtualHost),
		dropped:      make(map[string]interface{}),
		resources:    make(map[string]*typesv1.RoutingPolicy),
		virtualHosts: make(map[string]*virtualHostExt),
	}
}

func (r *routing) take() (map[string]types.Resource, []string) {

	updated := make(map[string]types.Resource, 0)
	dropped := make([]string, 0)

	for name, vh := range r.updated {
		updated[name] = vh
	}

	for name := range r.dropped {
		dropped = append(dropped, name)
	}

	r.updated = make(map[string]*routev3.VirtualHost)
	r.dropped = make(map[string]interface{})

	return updated, dropped

}

func (r *routing) dropPolicy(resourceId string) {
	delete(r.resources, resourceId)
}

func (r *routing) putPolicy(resourceId string, policy *typesv1.RoutingPolicy) {
	r.resources[resourceId] = policy
}

func (r *routing) processChanges() {

	vhxMap := map[string]*virtualHostExt{}

	resourceIds := sortStringKeys(r.resources)

	for _, resourceId := range resourceIds {

		policy := r.resources[resourceId]

		name := strings.Join([]string{policy.IngressName, strcase.ToKebab(policy.Domain)}, "/")

		if _, ok := vhxMap[name]; !ok {
			vhxMap[name] = &virtualHostExt{
				VirtualHost: &routev3.VirtualHost{
					Name:    name,
					Domains: []string{policy.Domain},
					Routes:  []*routev3.Route{},
				},
				routeMap: make(map[string]interface{}),
			}
		}

		vhx := vhxMap[name]

		for _, route := range policy.Routes {

			routeName := hashRoute(route, policy.IngressName, policy.Domain)

			if _, ok := vhx.routeMap[routeName]; ok {
				continue
			}
			vhx.routeMap[routeName] = nil

			envoyRoute := toEnvoyRoute(route)
			envoyRoute.Name = routeName

			vhx.Routes = append(vhx.Routes, envoyRoute)

		}

		vhx.hash()

	}

	for name, vh := range r.virtualHosts {

		vhx, ok := vhxMap[name]
		if ok {
			if !vhx.isEqual(vh) {
				r.updated[name] = vhx.VirtualHost
			}
			continue
		}

		r.dropped[name] = nil

	}

	for name, vhx := range vhxMap {

		if _, ok := r.virtualHosts[name]; ok {
			continue
		}

		r.updated[name] = vhx.VirtualHost

	}

	r.virtualHosts = vhxMap
}
