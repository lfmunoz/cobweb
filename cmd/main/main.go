package main

import (
	"github.com/lfmunoz/cobweb/internal/envoy"
	"github.com/lfmunoz/cobweb/internal/web"

	// LOGGING
	log "github.com/sirupsen/logrus"
)

// ________________________________________________________________________________
// MAIN
// ________________________________________________________________________________
func main() {
	log.Infof("[Main] - Starting Application...")
	go envoy.Start()
	web.Start()
}

/*
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
*/
