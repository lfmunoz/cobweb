package instance

import (
	"fmt"
	"sync"
)

type Instance struct {
	Id         int64
	NodeId     string
	Address    string
	ServiceIds []string
	Active     bool
}

// var concurrentMap map[string]InstanceInfo
// var instanceInfoMutex sync.RWMutex

var ConcurrentMap sync.Map

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
