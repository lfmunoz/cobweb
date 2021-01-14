package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	// LOGGING

	"github.com/lfmunoz/cobweb/internal/config"
	"github.com/lfmunoz/cobweb/internal/envoy"
	"github.com/lfmunoz/cobweb/internal/instance"
	log "github.com/sirupsen/logrus"
)

// ________________________________________________________________________________
// INSTANCE
// ________________________________________________________________________________
func getInstances(w http.ResponseWriter, req *http.Request) {
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

	before, _ := instance.LoadById(obj.Id)
	obj.Version = before.Version + 1
	log.Infof("[HTTP] - Before: %v ", before)
	log.Infof("[HTTP] - After: %v ", obj)
	var l = envoy.BuildListenerResource(obj.Local, obj.Remote)
	var c = envoy.BuildClusterResource(obj.Remote)
	instance.SendConfiguration(&obj, l, c)
	instance.Save(obj)
}

func importInfrastructure(w http.ResponseWriter, r *http.Request) {
	var obj []instance.Infrastructure
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

	log.Infof("[INFRA] - %v ", obj)

	/*
		before, _ := instance.LoadById(obj.Id)
		obj.Version = before.Version + 1
		log.Infof("[HTTP] - Before: %v ", before)
		log.Infof("[HTTP] - After: %v ", obj)
		var l = []types.Resource{
			config.BuildListenerResource(obj.Local, obj.Remote),
		}
		var c = []types.Resource{
			config.BuildClusterResource(obj.Remote),
		}
		instance.SendConfiguration(&obj, l, c)
		instance.Save(obj)
	*/
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
	http.HandleFunc("/api/importInfra", saveInstance)

	// http://localhost:8090/ will server index.html
	http.Handle("/", http.FileServer(http.Dir(appConfig.HttpDir)))

	//  nil tells it to use the default router
	http.ListenAndServe(listenBinding, nil)

}
