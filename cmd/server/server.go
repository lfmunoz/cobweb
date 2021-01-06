package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"google.golang.org/grpc"

	// LOGGING
	"github.com/lfmunoz/cobweb/internal/config"
	log "github.com/sirupsen/logrus"
)

const grpcMaxConcurrentStreams = 1000000

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

func pwd() {
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	fmt.Println(path)
}

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

// ________________________________________________________________________________
// MAIN
// ________________________________________________________________________________

func main() {
	pwd()

	local := config.Local{
		Name:    "local",
		Port:    8080,
		Address: "0.0.0.0",
	}

	remote := config.Remote{
		Name:    "google",
		Port:    80,
		Address: "google.com",
	}

	var l = []types.Resource{
		config.BuildListenerResource(local, remote),
	}
	var c = []types.Resource{
		config.BuildClusterResource(remote),
	}

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

	log.Infof(">>>>>>>>>>>>>>>>>>> creating snapshot Version " + fmt.Sprint(version))

	nodeId := cache.GetStatusKeys()[0]

	snap := cachev3.NewSnapshot(fmt.Sprint(version), nil, c, nil, l, nil, nil)
	if err := snap.Consistent(); err != nil {
		log.Errorf("snapshot inconsistency: %+v\n%+v", snap, err)
		os.Exit(1)
	}
	err := cache.SetSnapshot(nodeId, snap)
	if err != nil {
		log.Fatalf("Could not set snapshot %v", err)
	}

	time.Sleep(600 * time.Second)

}
