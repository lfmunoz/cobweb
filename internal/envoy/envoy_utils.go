package envoy

import (
	"time"

	// LOGGING
	"github.com/golang/protobuf/ptypes"
	"github.com/lfmunoz/cobweb/internal/instance"
	log "github.com/sirupsen/logrus"

	// GO CONTROL PLANE
	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	tcp_proxy "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
)

// ________________________________________________________________________________
// CLUSTER
// ________________________________________________________________________________
func BuildClusterResource(remote []instance.Remote) []types.Resource {
	var resource = []types.Resource{}
	for i := 1; i < len(remote); i++ {
		cluster := BuildCluster(remote[i])
		resource = append(resource, cluster)
	}
	return resource
}

func BuildCluster(remote instance.Remote) *cluster.Cluster {

	hst := &core.Address{Address: &core.Address_SocketAddress{
		SocketAddress: &core.SocketAddress{
			Address:  remote.Address,
			Protocol: core.SocketAddress_TCP,
			PortSpecifier: &core.SocketAddress_PortValue{
				PortValue: remote.Port,
			},
		},
	}}

	cluster := cluster.Cluster{
		Name:                 remote.Name, // netcat_cluster
		ConnectTimeout:       ptypes.DurationProto(2 * time.Second),
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_LOGICAL_DNS},
		DnsLookupFamily:      cluster.Cluster_V4_ONLY,
		LbPolicy:             cluster.Cluster_ROUND_ROBIN,
		LoadAssignment: &endpoint.ClusterLoadAssignment{
			ClusterName: remote.Name, // netcat_cluster
			Endpoints: []*endpoint.LocalityLbEndpoints{{
				LbEndpoints: []*endpoint.LbEndpoint{
					{
						HostIdentifier: &endpoint.LbEndpoint_Endpoint{
							Endpoint: &endpoint.Endpoint{
								Address: hst,
							}},
					},
				},
			}},
		},
	}

	return &cluster
}

// ________________________________________________________________________________
// LISTENER
// ________________________________________________________________________________
func BuildListenerResource(lis []instance.Local, cluster []instance.Remote) []types.Resource {
	var listeners = []types.Resource{}
	for i := 1; i < len(lis); i++ {
		resource := BuildListener(lis[i], cluster[i])
		listeners = append(listeners, resource)
	}
	return listeners

}

func BuildListener(lis instance.Local, cluster instance.Remote) *listener.Listener {

	log.Infof("[Creating listener] - %s", lis.Name)

	// https://github.com/envoyproxy/go-control-plane/blob/master/envoy/extensions/filters/network/tcp_proxy/v3/tcp_proxy.pb.go
	tcp := &tcp_proxy.TcpProxy{
		StatPrefix: "ingress_http",
		ClusterSpecifier: &tcp_proxy.TcpProxy_Cluster{
			Cluster: cluster.Name,
		},
	}
	pbst, err := ptypes.MarshalAny(tcp)
	if err != nil {
		log.Fatal(err)
	}

	listener := listener.Listener{
		Name: lis.Name,
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address:  lis.Address,
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: lis.Port,
					},
				},
			},
		},
		FilterChains: []*listener.FilterChain{{
			// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/listener/v3/listener_components.proto#config-listener-v3-filter
			Filters: []*listener.Filter{{
				// The name of the filter to instantiate. The name must match a supported filter.
				//  https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/network_filters/network_filters#config-network-filters
				//  https://github.com/envoyproxy/go-control-plane/blob/master/pkg/wellknown/wellknown.go
				Name: wellknown.TCPProxy,
				ConfigType: &listener.Filter_TypedConfig{
					TypedConfig: pbst,
				},
			}},
		}},
	}

	return &listener
}
