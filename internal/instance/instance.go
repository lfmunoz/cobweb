package instance

import (
	"fmt"
	"os"
	"sync"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/lfmunoz/cobweb/internal/config"

	// LOGGING
	log "github.com/sirupsen/logrus"
)

type Instance struct {
	Id      int64
	NodeId  string
	Address string
	// ServiceIds []string
	// Active     bool
	Version int32
	Local   config.Local
	Remote  config.Remote
}

// var concurrentMap map[string]InstanceInfo
// var instanceInfoMutex sync.RWMutex

var ConcurrentMap sync.Map

var Cache cachev3.SnapshotCache = cachev3.NewSnapshotCache(true, cachev3.IDHash{}, nil)

// ________________________________________________________________________________
// Instance endpoints
// ________________________________________________________________________________

func handler(key, value interface{}) bool {
	fmt.Printf("Name :%s %s\n", key, value)
	return true
}

func All() []Instance {
	instances := []Instance{}
	// Traversing, passing in a function, when the function is traversed,
	//  the function returns false to stop traversing
	ConcurrentMap.Range(func(key, value interface{}) bool {
		inst := value.(Instance)
		instances = append(instances, inst)
		return true
	})
	return instances
}

func Save(inst Instance) {
	ConcurrentMap.Store(inst.Id, inst)
}

func LoadById(id int64) (*Instance, bool) {
	var inst Instance
	result, ok := ConcurrentMap.Load(id)
	if ok {
		inst = result.(Instance)
		return &inst, ok
	} else {
		return nil, ok
	}
}

func LoadByNodeId(nodeId string) (*Instance, bool) {
	var inst Instance
	var found bool = false
	ConcurrentMap.Range(func(key, value interface{}) bool {
		inst = value.(Instance)
		if inst.NodeId == nodeId {
			found = true
			return false
		} else {
			return true
		}
	})
	return &inst, found
}

func DeleteById(id int64) {
	ConcurrentMap.Delete(id)
}

func BuildDefault(id int64, addr string) *Instance {
	return &Instance{
		Id:      id,
		NodeId:  "",
		Address: addr,
		Version: 1,
		Local: config.Local{
			Name:    "defaultListen",
			Address: "0.0.0.0",
			Port:    8080,
		},
		Remote: config.Remote{
			Name:    "defaultProxy",
			Address: "lfmunoz.com",
			Port:    80,
		},
	}
}

func SendConfiguration(inst *Instance, l []types.Resource, c []types.Resource) {
	snap := cachev3.NewSnapshot(fmt.Sprint(inst.Version), nil, c, nil, l, nil, nil)
	if err := snap.Consistent(); err != nil {
		log.Errorf("snapshot inconsistency: %+v\n%+v", snap, err)
		os.Exit(1)
	}
	err := Cache.SetSnapshot(inst.NodeId, snap)
	if err != nil {
		log.Fatalf("Could not set snapshot %v", err)
	}
}
