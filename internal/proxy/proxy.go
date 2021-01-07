package proxy

import (
	"fmt"
	"sync"
)

type Proxy struct {
	Id            int64
	NodeId        string
	ServiceId     string
	ListenAddress uint32
	ListenPort    uint32
	ProxyPort     uint32
	ProxyAddress  string
	// public_ip  string
	// private_ip string
	Dependencies []string
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

func All() []Proxy {
	instances := []Proxy{}
	// Traversing, passing in a function, when the function is traversed,
	//  the function returns false to stop traversing
	ConcurrentMap.Range(func(key, value interface{}) bool {
		inst := value.(Proxy)
		instances = append(instances, inst)
		return true
	})
	return instances
}

func Save(object Proxy) {
	ConcurrentMap.Store(object.Id, object)
}

func LoadById(id int64) (*Proxy, bool) {
	var inst Proxy
	result, ok := ConcurrentMap.Load(id)
	if ok {
		inst = result.(Proxy)
		return &inst, ok
	} else {
		return nil, ok
	}
}

func LoadByNodeId(nodeId string) (*Proxy, bool) {
	var inst Proxy
	var found bool = false
	ConcurrentMap.Range(func(key, value interface{}) bool {
		inst = value.(Proxy)
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
