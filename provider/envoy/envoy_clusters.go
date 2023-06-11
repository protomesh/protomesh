package envoy

import (
	"strings"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	typesv1 "github.com/protomesh/protomesh/proto/api/types/v1"
)

type envoyClusters struct {
	// map[resourceId]map[clusterName]
	Map map[string]map[string]interface{}
}

func newEnvoyClusters() *envoyClusters {
	return &envoyClusters{
		Map: make(map[string]map[string]interface{}),
	}
}

func (i *envoyClusters) Add(resourceId, clusterName string) {

	resource, ok := i.Map[resourceId]
	if !ok {
		resource = map[string]interface{}{}
	}
	i.Map[resourceId] = resource

	resource[clusterName] = nil

}

func (i *envoyClusters) Delete(resourceId string) {

	delete(i.Map, resourceId)

}

// Service = Cluster
type clustersEndpoints struct {
	// map[clusterName]map[region]map[zone]
	Map map[string]map[string]map[string][]*endpointv3.LbEndpoint
}

func newClustersEndpoints() *clustersEndpoints {
	return &clustersEndpoints{
		Map: make(map[string]map[string]map[string][]*endpointv3.LbEndpoint),
	}
}

func (c *clustersEndpoints) Add(clusterName string, instances ...*typesv1.InstanceSet_Instance) {

	cluster, ok := c.Map[clusterName]
	if !ok {
		cluster = map[string]map[string][]*endpointv3.LbEndpoint{}
	}
	c.Map[clusterName] = cluster

	for _, inst := range instances {

		region, ok := cluster[inst.Region]
		if !ok {
			region = map[string][]*endpointv3.LbEndpoint{}
		}
		cluster[inst.Region] = region

		zone, ok := region[inst.Zone]
		if !ok {
			zone = []*endpointv3.LbEndpoint{}
		}

		protocol := corev3.SocketAddress_TCP

		switch inst.TransportProtocol {

		case typesv1.TransportProtocol_TRANSPORT_PROTOCOL_UDP:
			protocol = corev3.SocketAddress_UDP

		}

		zone = append(zone, &endpointv3.LbEndpoint{
			HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
				Endpoint: &endpointv3.Endpoint{
					Hostname: inst.Hostname,
					Address: &corev3.Address{
						Address: &corev3.Address_SocketAddress{
							SocketAddress: &corev3.SocketAddress{
								Protocol: protocol,
								Address:  inst.Address,
								PortSpecifier: &corev3.SocketAddress_PortValue{
									PortValue: uint32(inst.Port),
								},
							},
						},
					},
				},
			},
		})

		region[inst.Zone] = zone

	}

}

func (c *clustersEndpoints) ToLoadAssignment(clusterName string) *endpointv3.ClusterLoadAssignment {

	cluster, ok := c.Map[clusterName]
	if !ok {
		return nil
	}

	load := &endpointv3.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints:   []*endpointv3.LocalityLbEndpoints{},
	}

	for region, zoneMap := range cluster {
		for zone, endpoints := range zoneMap {

			var locality *corev3.Locality

			if len(region) > 0 {

				locality = &corev3.Locality{
					Region: region,
				}

				if len(zone) > 0 {
					locality.Zone = zone
				}

			}

			load.Endpoints = append(load.Endpoints, &endpointv3.LocalityLbEndpoints{
				Locality:    locality,
				LbEndpoints: endpoints,
			})
		}

	}

	return load

}

func createClusterLoadId(clusterName string, resourceId string) string {
	return strings.Join([]string{clusterName, resourceId}, "-")
}
