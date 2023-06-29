package envoy

import (
	"strings"
	"time"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	corsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/cors/v3"
	grpcwebv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/grpc_web/v3"
	healthcheckv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/health_check/v3"
	jwtauthnv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/jwt_authn/v3"
	routerv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	http_connection_managerv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	httpv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	typesv1 "github.com/protomesh/protomesh/proto/api/types/v1"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func fromHttpIngress(node *typesv1.HttpIngress) (*listenerv3.Listener, *routev3.RouteConfiguration, error) {

	httpConn := &http_connection_managerv3.HttpConnectionManager{
		HttpFilters: []*http_connection_managerv3.HttpFilter{},
		StatPrefix:  node.IngressName,
		RouteSpecifier: &http_connection_managerv3.HttpConnectionManager_Rds{
			Rds: &http_connection_managerv3.Rds{
				ConfigSource: &corev3.ConfigSource{
					InitialFetchTimeout: durationpb.New(0 * time.Second),
					ResourceApiVersion:  resourcev3.DefaultAPIVersion,
					ConfigSourceSpecifier: &corev3.ConfigSource_ApiConfigSource{
						ApiConfigSource: &corev3.ApiConfigSource{
							ApiType:                   corev3.ApiConfigSource_GRPC,
							TransportApiVersion:       resourcev3.DefaultAPIVersion,
							SetNodeOnFirstMessageOnly: false,
							GrpcServices: []*corev3.GrpcService{{
								TargetSpecifier: &corev3.GrpcService_EnvoyGrpc_{
									EnvoyGrpc: &corev3.GrpcService_EnvoyGrpc{
										ClusterName: node.XdsClusterName,
									},
								},
							}},
						},
					},
				},
				RouteConfigName: node.IngressName,
			},
		},
	}

	for _, httpFilter := range node.HttpFilters {

		switch httpFilter := httpFilter.Filter.(type) {

		case *typesv1.HttpFilter_JwtAuthn_:

			jwtAuthn := &jwtauthnv3.JwtAuthentication{
				Providers: make(map[string]*jwtauthnv3.JwtProvider),
				Rules:     make([]*jwtauthnv3.RequirementRule, 0),
			}

			for _, provider := range httpFilter.JwtAuthn.Providers {

				jwtProvider := &jwtauthnv3.JwtProvider{
					Issuer:         provider.Issuer,
					Audiences:      provider.Audiences,
					Forward:        provider.Forward,
					ClaimToHeaders: []*jwtauthnv3.JwtClaimToHeader{},
					FromHeaders:    []*jwtauthnv3.JwtHeader{},
				}

				if provider.RemoteJwks != nil {
					jwtProvider.JwksSourceSpecifier = &jwtauthnv3.JwtProvider_RemoteJwks{
						RemoteJwks: &jwtauthnv3.RemoteJwks{
							HttpUri: &corev3.HttpUri{
								Uri: provider.RemoteJwks.HttpUri,
								HttpUpstreamType: &corev3.HttpUri_Cluster{
									Cluster: provider.RemoteJwks.ClusterName,
								},
								Timeout: provider.RemoteJwks.Timeout,
							},
						},
					}
				}

				for _, claimToHeader := range provider.ClaimToHeaders {

					jwtProvider.ClaimToHeaders = append(jwtProvider.ClaimToHeaders, &jwtauthnv3.JwtClaimToHeader{
						HeaderName: claimToHeader.HeaderName,
						ClaimName:  claimToHeader.ClaimName,
					})

				}

				for _, fromHeader := range provider.FromHeaders {

					jwtProvider.FromHeaders = append(jwtProvider.FromHeaders, &jwtauthnv3.JwtHeader{
						Name:        fromHeader.HeaderName,
						ValuePrefix: fromHeader.ValuePrefix,
					})

				}

				jwtAuthn.Providers[provider.ProviderName] = jwtProvider

			}

			for _, rule := range httpFilter.JwtAuthn.Rules {

				reqs := []*jwtauthnv3.JwtRequirement{}

				for _, providerName := range rule.RequiredProvidersNames {
					reqs = append(reqs, &jwtauthnv3.JwtRequirement{
						RequiresType: &jwtauthnv3.JwtRequirement_ProviderName{
							ProviderName: providerName,
						},
					})
				}

				jwtAuthn.Rules = append(jwtAuthn.Rules, &jwtauthnv3.RequirementRule{
					Match: &routev3.RouteMatch{
						PathSpecifier: &routev3.RouteMatch_Prefix{
							Prefix: rule.MatchPrefix,
						},
						CaseSensitive: wrapperspb.Bool(false),
					},
					RequirementType: &jwtauthnv3.RequirementRule_Requires{
						Requires: &jwtauthnv3.JwtRequirement{
							RequiresType: &jwtauthnv3.JwtRequirement_RequiresAny{
								RequiresAny: &jwtauthnv3.JwtRequirementOrList{
									Requirements: reqs,
								},
							},
						},
					},
				})

			}

			jwtAuthnAny, _ := anypb.New(jwtAuthn)

			httpConn.HttpFilters = append(httpConn.HttpFilters, &http_connection_managerv3.HttpFilter{
				Name: strings.Join([]string{node.IngressName, "jwtAuthn"}, "-"),
				ConfigType: &http_connection_managerv3.HttpFilter_TypedConfig{
					TypedConfig: jwtAuthnAny,
				},
			})

		case *typesv1.HttpFilter_HealthCheck_:

			healthCheckAny, _ := anypb.New(&healthcheckv3.HealthCheck{
				PassThroughMode:              wrapperspb.Bool(false),
				ClusterMinHealthyPercentages: make(map[string]*typev3.Percent),
				Headers: []*routev3.HeaderMatcher{
					{
						Name: ":path",
						HeaderMatchSpecifier: &routev3.HeaderMatcher_ExactMatch{
							ExactMatch: httpFilter.HealthCheck.Path,
						},
					},
				},
			})

			httpConn.HttpFilters = append(httpConn.HttpFilters, &http_connection_managerv3.HttpFilter{
				Name: strings.Join([]string{node.IngressName, "healthCheck"}, "-"),
				ConfigType: &http_connection_managerv3.HttpFilter_TypedConfig{
					TypedConfig: healthCheckAny,
				},
			})

		case *typesv1.HttpFilter_GrpcWeb_:

			if !httpFilter.GrpcWeb.Enable {
				break
			}

			gprcWebAny, _ := anypb.New(&grpcwebv3.GrpcWeb{})

			httpConn.HttpFilters = append(httpConn.HttpFilters, &http_connection_managerv3.HttpFilter{
				Name: strings.Join([]string{node.IngressName, "grpcWeb"}, "-"),
				ConfigType: &http_connection_managerv3.HttpFilter_TypedConfig{
					TypedConfig: gprcWebAny,
				},
			})

		case *typesv1.HttpFilter_Cors_:

			if !httpFilter.Cors.Enable {
				break
			}

			corsAny, _ := anypb.New(&corsv3.Cors{})

			httpConn.HttpFilters = append(httpConn.HttpFilters, &http_connection_managerv3.HttpFilter{
				Name: strings.Join([]string{node.IngressName, "cors"}, "-"),
				ConfigType: &http_connection_managerv3.HttpFilter_TypedConfig{
					TypedConfig: corsAny,
				},
			})

		}

	}

	routerAny, _ := anypb.New(&routerv3.Router{})

	httpConn.HttpFilters = append(httpConn.HttpFilters, &http_connection_managerv3.HttpFilter{
		Name: strings.Join([]string{node.IngressName, "httpRouter"}, "-"),
		ConfigType: &http_connection_managerv3.HttpFilter_TypedConfig{
			TypedConfig: routerAny,
		},
	})

	httpConnAny, err := anypb.New(httpConn)
	if err != nil {
		return nil, nil, err
	}

	routeConf := &routev3.RouteConfiguration{
		Name:                     node.IngressName,
		IgnorePortInHostMatching: true,
		Vhds: &routev3.Vhds{
			ConfigSource: &corev3.ConfigSource{
				InitialFetchTimeout: durationpb.New(0 * time.Second),
				ResourceApiVersion:  resource.DefaultAPIVersion,
				ConfigSourceSpecifier: &corev3.ConfigSource_ApiConfigSource{
					ApiConfigSource: &corev3.ApiConfigSource{
						ApiType:                   corev3.ApiConfigSource_DELTA_GRPC,
						TransportApiVersion:       resource.DefaultAPIVersion,
						SetNodeOnFirstMessageOnly: true,
						GrpcServices: []*corev3.GrpcService{{
							TargetSpecifier: &corev3.GrpcService_EnvoyGrpc_{
								EnvoyGrpc: &corev3.GrpcService_EnvoyGrpc{
									ClusterName: node.XdsClusterName,
								},
							},
						}},
					},
				},
			},
		},
	}

	return &listenerv3.Listener{
		Name: node.IngressName,
		Address: &corev3.Address{
			Address: &corev3.Address_SocketAddress{
				SocketAddress: &corev3.SocketAddress{
					Protocol: corev3.SocketAddress_TCP,
					Address:  "0.0.0.0",
					PortSpecifier: &corev3.SocketAddress_PortValue{
						PortValue: uint32(node.ListenPort),
					},
				},
			},
		},
		FilterChains: []*listenerv3.FilterChain{
			{
				Filters: []*listenerv3.Filter{
					{
						Name: wellknown.HTTPConnectionManager,
						ConfigType: &listenerv3.Filter_TypedConfig{
							TypedConfig: httpConnAny,
						},
					},
				},
			},
		},
	}, routeConf, nil
}

func fromService(node *typesv1.Service) (*clusterv3.Cluster, error) {

	cluster := &clusterv3.Cluster{
		Name: node.ServiceName,
		ClusterDiscoveryType: &clusterv3.Cluster_Type{
			// Type: clusterv3.Cluster_STATIC,
			Type: clusterv3.Cluster_EDS,
		},
		EdsClusterConfig: &clusterv3.Cluster_EdsClusterConfig{
			ServiceName: node.ServiceName,
			EdsConfig: &corev3.ConfigSource{
				ResourceApiVersion: resourcev3.DefaultAPIVersion,
				ConfigSourceSpecifier: &corev3.ConfigSource_ApiConfigSource{
					ApiConfigSource: &corev3.ApiConfigSource{
						ApiType:                   corev3.ApiConfigSource_GRPC,
						TransportApiVersion:       resourcev3.DefaultAPIVersion,
						SetNodeOnFirstMessageOnly: false,
						GrpcServices: []*corev3.GrpcService{{
							TargetSpecifier: &corev3.GrpcService_EnvoyGrpc_{
								EnvoyGrpc: &corev3.GrpcService_EnvoyGrpc{
									ClusterName: node.XdsClusterName,
								},
							},
						}},
					},
				},
			},
		},
		LbPolicy:                      clusterv3.Cluster_ROUND_ROBIN,
		DnsLookupFamily:               clusterv3.Cluster_AUTO,
		ConnectTimeout:                node.ConnectTimeout,
		TypedExtensionProtocolOptions: map[string]*anypb.Any{},
	}

	switch protoOpts := node.InstanceApplicationProtocolOptions.(type) {

	case *typesv1.Service_InstanceHttp1Options:

		protocolOptsAny, err := anypb.New(&httpv3.HttpProtocolOptions{
			UpstreamProtocolOptions: &httpv3.HttpProtocolOptions_ExplicitHttpConfig_{
				ExplicitHttpConfig: &httpv3.HttpProtocolOptions_ExplicitHttpConfig{
					ProtocolConfig: &httpv3.HttpProtocolOptions_ExplicitHttpConfig_HttpProtocolOptions{
						HttpProtocolOptions: &corev3.Http1ProtocolOptions{},
					},
				},
			},
		})
		if err != nil {
			return nil, err
		}

		cluster.TypedExtensionProtocolOptions["envoy.extensions.upstreams.http.v3.HttpProtocolOptions"] = protocolOptsAny

	case *typesv1.Service_InstanceHttp2Options:

		protocolOptsAny, err := anypb.New(&httpv3.HttpProtocolOptions{
			UpstreamProtocolOptions: &httpv3.HttpProtocolOptions_ExplicitHttpConfig_{
				ExplicitHttpConfig: &httpv3.HttpProtocolOptions_ExplicitHttpConfig{
					ProtocolConfig: &httpv3.HttpProtocolOptions_ExplicitHttpConfig_Http2ProtocolOptions{
						Http2ProtocolOptions: &corev3.Http2ProtocolOptions{
							MaxConcurrentStreams: wrapperspb.UInt32(uint32(protoOpts.InstanceHttp2Options.MaxConcurrentStreams)),
						},
					},
				},
			},
		})
		if err != nil {
			return nil, err
		}

		cluster.TypedExtensionProtocolOptions["envoy.extensions.upstreams.http.v3.HttpProtocolOptions"] = protocolOptsAny

	}

	return cluster, nil

}
