package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	// LOGGING

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/lfmunoz/cobweb/internal/config"
	"github.com/lfmunoz/cobweb/internal/envoy"
	"github.com/lfmunoz/cobweb/internal/instance"
	log "github.com/sirupsen/logrus"
)

// ________________________________________________________________________________
// INSTANCE
// ________________________________________________________________________________
func getInstances(w http.ResponseWriter, req *http.Request) {
	log.Infof("[HTTP] - get instances")
	instances := instance.All()
	w.Header().Set("Content-Type", "application/json")
	b, err := json.Marshal(instances)
	if err == nil {
		w.WriteHeader(200)
		w.Write(b)
	} else {
		w.WriteHeader(500)
	}
}

func saveInstance(w http.ResponseWriter, r *http.Request) {
	log.Infof("[HTTP] - save instance ")
	var obj instance.Instance
	err := decodeJSONBody(w, r, &obj)
	if err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			http.Error(w, mr.msg, mr.status)
		} else {
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	before, ok := instance.LoadByNodeId(obj.NodeId)
	if !ok {
		log.Errorf("[HTTP] - could not find: %s ", obj.NodeId)
		return
	}
	obj.Version = before.Version + 1
	log.Infof("[HTTP] - Before: %v ", before)
	log.Infof("[HTTP] - After: %v ", obj)

	listenerResource := []types.Resource{}
	clusterResource := []types.Resource{}
	for i := 0; i < len(obj.Local); i++ {
		l := envoy.BuildListener(obj.Local[i], obj.Remote[i])
		c := envoy.BuildCluster(obj.Remote[i])
		listenerResource = append(listenerResource, l)
		clusterResource = append(clusterResource, c)
	}
	instance.SendConfiguration(&obj, listenerResource, clusterResource)

	instance.Save(obj)
}

func importInstances(w http.ResponseWriter, r *http.Request) {
	log.Infof("[HTTP] - import infrastructure")
	var objs []instance.Instance
	err := decodeJSONBody(w, r, &objs)
	if err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			http.Error(w, mr.msg, mr.status)
		} else {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		log.Println(err.Error())
		return
	}

	log.Infof("[HTTP] - [IMPORTING]: %+v ", objs)
	for _, obj := range objs {
		before, ok := instance.LoadByNodeId(obj.NodeId)
		if ok {
			obj.Version = before.Version + 1
			log.Infof("[HTTP] - Before: %v ", before)
			log.Infof("[HTTP] - After: %v ", obj)
		} else {
			log.Infof("[HTTP] - New: %v ", obj)
		}
		var l = envoy.BuildListenerResource(obj.Local, obj.Remote)
		var c = envoy.BuildClusterResource(obj.Remote)
		instance.SendConfiguration(&obj, l, c)
		instance.Save(obj)
	}
}

func infraToInstances(objs []instance.Infrastructure, ipMap map[string]string) []instance.Instance {
	var insts []instance.Instance
	for _, obj := range objs {
		for i, remote := range obj.Remote {
			addr, ok := ipMap[remote.Address]
			if ok {
				obj.Remote[i].Address = addr
			}
		}

		inst := instance.Instance{
			Address:      obj.Public_ip,
			Local:        obj.Local,
			Remote:       obj.Remote,
			NodeId:       obj.Name,
			Version:      1,
			Dependencies: obj.Dependencies,
		}
		insts = append(insts, inst)

	}
	return insts
}

func importInfra(w http.ResponseWriter, r *http.Request) {
	log.Infof("[HTTP] - import infrastructure")

	ipMap := make(map[string]string)

	var objs []instance.Infrastructure
	err := decodeJSONBody(w, r, &objs)
	if err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			http.Error(w, mr.msg, mr.status)
		} else {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		log.Println(err.Error())
		return
	}

	for _, inf := range objs {
		ipMap[inf.Name] = inf.Public_ip
	}

	insts := infraToInstances(objs, ipMap)

	log.Infof("[HTTP] - [IMPORTING]: %+v ", objs)
	for _, obj := range insts {
		before, ok := instance.LoadByNodeId(obj.NodeId)
		if ok {
			obj.Version = before.Version + 1
			log.Infof("[HTTP] - Before: %v ", before)
			log.Infof("[HTTP] - After: %v ", obj)
		} else {
			log.Infof("[HTTP] - New: %v ", obj)
		}
		var l = envoy.BuildListenerResource(obj.Local, obj.Remote)
		var c = envoy.BuildClusterResource(obj.Remote)
		instance.SendConfiguration(&obj, l, c)
		instance.Save(obj)
	}

}

// ________________________________________________________________________________
// ENTRY
// ________________________________________________________________________________
func Start(appConfig config.AppConfig) {
	listenBinding := fmt.Sprintf(":%d", appConfig.HttpPort)
	log.WithFields(log.Fields{"addr": listenBinding}).Info("[HTTP] - http listening ")
	// INSTANCE
	http.HandleFunc("/api/instance", getInstances)
	http.HandleFunc("/api/saveInstance", saveInstance)
	http.HandleFunc("/api/importInstances", importInstances)
	http.HandleFunc("/api/importInfra", importInfra)

	// http://localhost:8090/ will server index.html
	http.Handle("/", http.FileServer(http.Dir(appConfig.HttpDir)))

	//  nil tells it to use the default router
	http.ListenAndServe(listenBinding, nil)

}
