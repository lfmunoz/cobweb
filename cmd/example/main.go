package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"

	"sync"
	"sync/atomic"
	"time"

	// LOGGING
	log "github.com/sirupsen/logrus"

	// GRPC

	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc"

	// GO CONTROL PLANE
	envoy_api_v2_auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
)

var (
	debug       bool
	onlyLogging bool
	withALS     bool

	localhost         = "0.0.0.0" //127.0.0.1
	listenPort uint32 = 80

	port        uint
	gatewayPort uint
	alsPort     uint

	mode string

	version int32

	cache cachev3.SnapshotCache

	strSlice = []string{"www.bbc.com", "www.yahoo.com", "blog.salrashid.me"}
)

const (
	XdsCluster = "xds_cluster"
	Ads        = "ads"
	Xds        = "xds"
	Rest       = "rest"
)

func init() {
	flag.BoolVar(&debug, "debug", true, "Use debug logging")
	flag.UintVar(&port, "port", 18000, "Management server port")
	flag.UintVar(&gatewayPort, "gateway", 18001, "Management server port for HTTP gateway")
	flag.StringVar(&mode, "ads", Ads, "Management server type (ads only now)")
}

// ________________________________________________________________________________
// callback handlers
// ________________________________________________________________________________
func (cb *Callbacks) Report() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	log.WithFields(log.Fields{"fetches": cb.Fetches, "requests": cb.Requests}).Info("cb.Report()  callbacks")
}
func (cb *Callbacks) OnStreamOpen(_ context.Context, id int64, typ string) error {
	log.Infof("OnStreamOpen %d open for %s", id, typ)
	return nil
}
func (cb *Callbacks) OnStreamClosed(id int64) {
	log.Infof("OnStreamClosed %d closed", id)
}
func (cb *Callbacks) OnStreamRequest(id int64, r *discoverygrpc.DiscoveryRequest) error {
	log.Infof("OnStreamRequest %v", r.TypeUrl)
	log.Infof("OnStreamRequest %v", r.Node.Id)
	log.Infof("OnStreamRequest %v", r.Node.Cluster)
	log.Infof("OnStreamRequest %v", r.Node.ListeningAddresses)
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.Requests++
	if cb.Signal != nil {
		close(cb.Signal)
		cb.Signal = nil
	}
	return nil
}
func (cb *Callbacks) OnStreamResponse(int64, *discoverygrpc.DiscoveryRequest, *discoverygrpc.DiscoveryResponse) {
	log.Infof("OnStreamResponse...")
	cb.Report()
}
func (cb *Callbacks) OnFetchRequest(ctx context.Context, req *discoverygrpc.DiscoveryRequest) error {
	log.Infof("OnFetchRequest...")
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.Fetches++
	if cb.Signal != nil {
		close(cb.Signal)
		cb.Signal = nil
	}
	return nil
}
func (cb *Callbacks) OnFetchResponse(*discoverygrpc.DiscoveryRequest, *discoverygrpc.DiscoveryResponse) {
	log.Infof("OnFetchResponse...")
}

type Callbacks struct {
	Signal   chan struct{}
	Debug    bool
	Fetches  int
	Requests int
	mu       sync.Mutex
}

// ________________________________________________________________________________
// RunManagementServer starts an xDS server at the given port.
// ________________________________________________________________________________
const grpcMaxConcurrentStreams = 1000000

func RunManagementServer(ctx context.Context, server serverv3.Server, port uint) {
	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions, grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams))
	grpcServer := grpc.NewServer(grpcOptions...)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.WithError(err).Fatal("failed to listen")
	}

	// register services
	discoverygrpc.RegisterAggregatedDiscoveryServiceServer(grpcServer, server)

	log.WithFields(log.Fields{"port": port}).Info("[Management Server Listening]")
	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			log.Error(err)
		}
	}()
	<-ctx.Done()

	grpcServer.GracefulStop()
}

func pwd() {
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	fmt.Println(path)
}

// ________________________________________________________________________________
// MAIN
// ________________________________________________________________________________
func main() {
	pwd()

	flag.Parse()
	log.SetLevel(log.DebugLevel)
	// A Context carries a deadline, cancelation signal, and request-scoped values
	// 	across API boundaries.
	ctx := context.Background()
	log.Printf("[Starting] - Control Plane Application")

	signal := make(chan struct{})
	cb := &Callbacks{
		Signal:   signal,
		Fetches:  0,
		Requests: 0,
	}
	cache = cachev3.NewSnapshotCache(true, cachev3.IDHash{}, nil)

	srv := serverv3.NewServer(ctx, cache, cb)

	// start the xDS server
	go RunManagementServer(ctx, srv, port)

	<-signal

	for _, v := range strSlice {

		nodeId := cache.GetStatusKeys()[0]

		// ________________________________________________________________________________
		// Create Cluster
		// ________________________________________________________________________________
		var clusterName = "service_bbc"
		var remoteHost = v

		log.Infof("[Creating Cluster] - remoteHost= %s, nodeID=%s,  clusterName=%s", v, nodeId, clusterName)

		hst := &core.Address{Address: &core.Address_SocketAddress{
			SocketAddress: &core.SocketAddress{
				Address:  remoteHost,
				Protocol: core.SocketAddress_TCP,
				PortSpecifier: &core.SocketAddress_PortValue{
					PortValue: uint32(443),
				},
			},
		}}
		uctx := &envoy_api_v2_auth.UpstreamTlsContext{}
		tctx, err := ptypes.MarshalAny(uctx)
		if err != nil {
			log.Fatal(err)
		}

		c := []types.Resource{
			&cluster.Cluster{
				Name:                 clusterName,
				ConnectTimeout:       ptypes.DurationProto(2 * time.Second),
				ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_LOGICAL_DNS},
				DnsLookupFamily:      cluster.Cluster_V4_ONLY,
				LbPolicy:             cluster.Cluster_ROUND_ROBIN,
				LoadAssignment: &endpoint.ClusterLoadAssignment{
					ClusterName: clusterName,
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
				TransportSocket: &core.TransportSocket{
					Name: "envoy.transport_sockets.tls",
					ConfigType: &core.TransportSocket_TypedConfig{
						TypedConfig: tctx,
					},
				},
			},
		}

		// ________________________________________________________________________________
		// Create Listener
		// ________________________________________________________________________________
		var listenerName = "listener_0"
		var targetHost = v

		log.Infof("[Creating listener] - %s", listenerName)

		rte := &route.RouteConfiguration{
			Name: "local_route",
			VirtualHosts: []*route.VirtualHost{{
				Name:    "local_service",
				Domains: []string{"*"},
				Routes: []*route.Route{{
					Match: &route.RouteMatch{
						PathSpecifier: &route.RouteMatch_Prefix{
							Prefix: "/",
						},
					},
					Action: &route.Route_Route{
						Route: &route.RouteAction{
							ClusterSpecifier: &route.RouteAction_Cluster{
								Cluster: clusterName,
							},
							PrefixRewrite: "/robots.txt",
							HostRewriteSpecifier: &route.RouteAction_HostRewriteLiteral{
								HostRewriteLiteral: targetHost,
							},
						},
					},
				}},
			}},
		}

		manager := &hcm.HttpConnectionManager{
			CodecType:  hcm.HttpConnectionManager_AUTO,
			StatPrefix: "ingress_http",
			RouteSpecifier: &hcm.HttpConnectionManager_RouteConfig{
				RouteConfig: rte,
			},
			HttpFilters: []*hcm.HttpFilter{{
				Name: wellknown.Router,
			}},
		}
		pbst, err := ptypes.MarshalAny(manager)
		if err != nil {
			log.Fatal(err)
		}

		var l = []types.Resource{
			&listener.Listener{
				Name: listenerName,
				Address: &core.Address{
					Address: &core.Address_SocketAddress{
						SocketAddress: &core.SocketAddress{
							Protocol: core.SocketAddress_TCP,
							Address:  localhost,
							PortSpecifier: &core.SocketAddress_PortValue{
								PortValue: listenPort,
							},
						},
					},
				},
				FilterChains: []*listener.FilterChain{{
					Filters: []*listener.Filter{{
						Name: wellknown.HTTPConnectionManager,
						ConfigType: &listener.Filter_TypedConfig{
							TypedConfig: pbst,
						},
					}},
					// TransportSocket: &core.TransportSocket{
					// Name: "envoy.transport_sockets.tls",
					// ConfigType: &core.TransportSocket_TypedConfig{
					// TypedConfig: scfg,
					// },
					// },
				}},
			}}

		// =================================================================================
		atomic.AddInt32(&version, 1)
		log.Infof(">>>>>>>>>>>>>>>>>>> creating snapshot Version " + fmt.Sprint(version))

		snap := cachev3.NewSnapshot(fmt.Sprint(version), nil, c, nil, l, nil, nil)
		if err := snap.Consistent(); err != nil {
			log.Errorf("snapshot inconsistency: %+v\n%+v", snap, err)
			os.Exit(1)
		}
		err = cache.SetSnapshot(nodeId, snap)
		if err != nil {
			log.Fatalf("Could not set snapshot %v", err)
		}

		//reader := bufio.NewReader(os.Stdin)
		//_, _ = reader.ReadString('\n')

		time.Sleep(60 * time.Second)

	}

}
