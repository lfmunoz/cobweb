package instance

import (
	"fmt"
	"os"
	"sync"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"

	// LOGGING
	log "github.com/sirupsen/logrus"
)

// ________________________________________________________________________________
// Data Structure
// ________________________________________________________________________________
type Local struct {
	Name    string
	Port    uint32
	Address string
}

type Remote struct {
	Name    string
	Port    uint32
	Address string
}

/*
type Infrastructure struct {
	Name         string
	Private_ip   string
	Public_ip    string
	Gateway      string
	Scripts      []string
	Dependencies []string
	Local        []Local
	Remote       []Remote
}
*/

type Instance struct {
	Id           int64
	NodeId       string
	Address      string
	Version      int32
	Local        []Local
	Remote       []Remote
	Dependencies []string
}

// var concurrentMap map[string]InstanceInfo
// var instanceInfoMutex sync.RWMutex

var ConcurrentMap sync.Map

var Cache cachev3.SnapshotCache = cachev3.NewSnapshotCache(true, cachev3.IDHash{}, nil)

func DtoToEntity() {

}

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
	ConcurrentMap.Store(inst.NodeId, inst)
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

func DeleteByNodeId(nodeId string) {
	ConcurrentMap.Delete(nodeId)
}

func DeleteById(id int64) {
	var inst Instance
	ConcurrentMap.Range(func(key, value interface{}) bool {
		inst = value.(Instance)
		if inst.Id == id {
			ConcurrentMap.Delete(inst.NodeId)
			return false
		} else {
			return true
		}
	})
}

func BuildDefault(id int64, addr string) Instance {
	return Instance{
		Id:      id,
		NodeId:  "",
		Address: addr,
		Version: 1,
		Local: []Local{{
			Name:    "default-local",
			Address: "0.0.0.0",
			Port:    8080,
		}},
		Remote: []Remote{{
			Name:    "default-remote",
			Address: "apache.org", // nginx.org
			Port:    80,
		}},
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
