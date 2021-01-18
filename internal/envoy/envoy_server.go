package envoy

import (
	"context"
	"fmt"
	"net"
	"sync"

	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/lfmunoz/cobweb/internal/config"
	"github.com/lfmunoz/cobweb/internal/instance"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"

	// LOGGING
	log "github.com/sirupsen/logrus"
)

// ________________________________________________________________________________
// GLOBAL
// ________________________________________________________________________________
type Callbacks struct {
	Signal   chan struct{}
	Debug    bool
	Fetches  int
	Requests int
	mu       sync.Mutex
	Cache    cachev3.SnapshotCache
}

// ________________________________________________________________________________
// CONFIG
// ________________________________________________________________________________
const grpcMaxConcurrentStreams = 1000

var (
	debug       bool
	onlyLogging bool
	withALS     bool
	mode        string
	version     int32
	cache       cachev3.SnapshotCache
	connections sync.Map
)

type Connection struct {
	Addr         string
	NodeId       string
	ConnectionId int64
}

// ________________________________________________________________________________
// callback handlers
// ________________________________________________________________________________
func (cb *Callbacks) Report() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	log.WithFields(log.Fields{"fetches": cb.Fetches, "requests": cb.Requests}).Info("cb.Report()  callbacks")
}

func (cb *Callbacks) OnStreamOpen(ctx context.Context, id int64, typ string) error {
	p, _ := peer.FromContext(ctx)
	addr := p.Addr.String()

	connection := Connection{addr, "", id}
	connections.Store(id, connection)

	log.Infof("[Envoy]-[OnStreamOpen - %d] - typ=%s addr=%s", id, typ, addr)
	return nil
}

func (cb *Callbacks) OnStreamClosed(id int64) {
	log.Infof("[Envoy]-[OnStreamClosed - %d] - closed", id)
	connections.Delete(id)
}

func (cb *Callbacks) OnStreamRequest(id int64, r *discoverygrpc.DiscoveryRequest) error {
	// update connection
	connectionResult, ok := connections.Load(id)
	if !ok {
		log.Errorf("[Envoy]-[OnStreamRequest - %d] - connection not found", id)
	}
	connection := connectionResult.(Connection)
	connection.NodeId = r.Node.Id
	connections.Store(id, connection)
	log.Infof("[Envoy]-[OnStreamRequest - %d] - nodeId=%s addr=%s",
		id, connection.NodeId, connection.Addr)

	// update instance
	var obj instance.Instance
	result, ok := instance.LoadByNodeId(connection.NodeId)
	if !ok {
		obj = instance.BuildDefault(connection.ConnectionId, connection.Addr)
		obj.NodeId = connection.NodeId
	} else {
		obj = *result
		obj.Id = connection.ConnectionId
		obj.Address = connection.Addr
		obj.NodeId = connection.NodeId
	}
	instance.Save(obj)
	log.Infof("[Envoy]-[Sending Config] - %+v", obj)

	listenerResource := []types.Resource{}
	clusterResource := []types.Resource{}
	for i := 0; i < len(obj.Local); i++ {
		l := BuildListener(obj.Local[i], obj.Remote[i])
		c := BuildCluster(obj.Remote[i])
		listenerResource = append(listenerResource, l)
		clusterResource = append(clusterResource, c)
	}
	instance.SendConfiguration(&obj, listenerResource, clusterResource)

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
	log.Infof("[Envoy] - OnStreamResponse...")
	cb.Report()
}

func (cb *Callbacks) OnFetchRequest(ctx context.Context, req *discoverygrpc.DiscoveryRequest) error {
	log.Infof("[Envoy] - OnFetchRequest...")
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
	log.Infof("[Envoy] - OnFetchResponse...")
}

// ________________________________________________________________________________
// RunManagementServer starts an xDS server at the given port.
// ________________________________________________________________________________
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

	log.WithFields(log.Fields{"port": port}).Info("[Envoy] - gRPC listening ")
	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			log.Error(err)
		}
	}()
	<-ctx.Done()

	grpcServer.GracefulStop()
}

func Start(appConfig config.AppConfig) {

	// A Context carries a deadline, cancelation signal, and request-scoped values
	// 	across API boundaries.
	ctx := context.Background()
	log.Infof("[Envoy] - Control Plane Application Initializing...")

	signal := make(chan struct{})
	cb := &Callbacks{
		Signal:   signal,
		Fetches:  0,
		Requests: 0,
		Cache:    instance.Cache,
	}

	srv := serverv3.NewServer(ctx, cb.Cache, cb)

	// start the xDS server
	go RunManagementServer(ctx, srv, appConfig.GrpcPort)

	<-signal
}
