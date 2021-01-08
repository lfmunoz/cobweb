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

var tempConnectionMap sync.Map

// ________________________________________________________________________________
// CONFIG
// ________________________________________________________________________________
const grpcMaxConcurrentStreams = 1000

var (
	port uint = 18000

	debug       bool
	onlyLogging bool
	withALS     bool
	mode        string
	version     int32
	cache       cachev3.SnapshotCache
)

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

	result, ok := instance.LoadById(id)
	if ok {
		result.Address = addr
		// result.Active = true
		instance.Save(*result)
	} else {
		instance.Save(*instance.BuildDefault(id, addr))
	}

	log.Infof("[Envoy]-[OnStreamOpen - %d] - typ=%s addr=%s", id, typ, addr)
	return nil
}

func (cb *Callbacks) OnStreamClosed(id int64) {
	log.Infof("[Envoy]-[OnStreamClosed - %d] - closed", id)
	instance.DeleteById(id)
}

func (cb *Callbacks) OnStreamRequest(id int64, r *discoverygrpc.DiscoveryRequest) error {
	result, ok := instance.LoadById(id)
	if !ok {
		log.Errorf("[Envoy]-[OnStreamRequest - %d] - item not found", id)
	} else {
		result.NodeId = r.Node.Id
		instance.Save(*result)
	}
	log.Infof("[Envoy]-[OnStreamRequest - %d] - nodeId=%v addr=%s cluster=%s",
		id, r.Node.Id, result.Address, r.Node.Cluster)
	// log.Infof("[Envoy] - OnStreamRequest %v", r.TypeUrl)
	// log.Infof("OnStreamRequest %v", r.Node.Id)
	// log.Infof("OnStreamRequest %v", r.Node.Cluster)
	// log.Infof("OnStreamRequest %v", r.Node.ListeningAddresses)
	// log.Infof("OnStreamRequest %v", r.ResourceNames)

	var l = []types.Resource{
		config.BuildListenerResource(result.Local, result.Remote),
	}
	var c = []types.Resource{
		config.BuildClusterResource(result.Remote),
	}

	instance.SendConfiguration(result, l, c)
	/*
		atomic.AddInt32(&result.Version, 1)
		snap := cachev3.NewSnapshot(fmt.Sprint(result.Version), nil, c, nil, l, nil, nil)
		if err := snap.Consistent(); err != nil {
			log.Errorf("snapshot inconsistency: %+v\n%+v", snap, err)
			os.Exit(1)
		}
		err := cb.Cache.SetSnapshot(result.NodeId, snap)
		if err != nil {
			log.Fatalf("Could not set snapshot %v", err)
		}
	*/

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

func Start() {

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
		// Cache:    cachev3.NewSnapshotCache(true, cachev3.IDHash{}, nil),
	}
	// cache = cachev3.NewSnapshotCache(true, cachev3.IDHash{}, nil)

	srv := serverv3.NewServer(ctx, cb.Cache, cb)

	// start the xDS server
	go RunManagementServer(ctx, srv, port)

	<-signal
}
