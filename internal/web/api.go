package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	// LOGGING

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/lfmunoz/cobweb/internal/config"
	"github.com/lfmunoz/cobweb/internal/instance"
	"github.com/lfmunoz/cobweb/internal/proxy"
	log "github.com/sirupsen/logrus"
)

// ________________________________________________________________________________
// CONFIG
// ________________________________________________________________________________
const httpPort = ":8090"

// ________________________________________________________________________________
// ENDPOINTS
// ________________________________________________________________________________
/*
func Hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "hello\n")
}

func Headers(w http.ResponseWriter, req *http.Request) {
	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

func Context(w http.ResponseWriter, req *http.Request) {
	// A context.Context is created for each request by the net/http machinery,
	//  and is available with the Context() method.
	ctx := req.Context()
	fmt.Println("server: hello handler started")
	defer fmt.Println("server: hello handler ended")

	select {
	case <-time.After(5 * time.Second):
		fmt.Fprintf(w, "hello\n")
	case <-ctx.Done():

		err := ctx.Err()
		fmt.Println("server:", err)
		internalError := http.StatusInternalServerError
		http.Error(w, err.Error(), internalError)
	}
}
*/

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
	var l = []types.Resource{
		config.BuildListenerResource(obj.Local, obj.Remote),
	}
	var c = []types.Resource{
		config.BuildClusterResource(obj.Remote),
	}
	instance.SendConfiguration(&obj, l, c)
	instance.Save(obj)
}

// ________________________________________________________________________________
// PROXY
// ________________________________________________________________________________
func saveProxyById(w http.ResponseWriter, r *http.Request) {
	var p proxy.Proxy
	err := decodeJSONBody(w, r, &p)
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
	proxy.Save(p)
}

func saveProxyArray(w http.ResponseWriter, r *http.Request) {
	var p []proxy.Proxy
	err := decodeJSONBody(w, r, &p)
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
	fmt.Println(p)
}

func getProxies(w http.ResponseWriter, req *http.Request) {
	objects := proxy.All()
	w.Header().Set("Content-Type", "application/json")
	b, err := json.Marshal(objects)
	if err == nil {
		w.WriteHeader(200)
		w.Write(b)
	} else {
		w.WriteHeader(500)
	}
}

// ________________________________________________________________________________
// ENTRY
// ________________________________________________________________________________
func Start() {
	log.WithFields(log.Fields{"port": httpPort}).Info("[HTTP] - http listening ")
	// INSTANCE
	http.HandleFunc("/api/instance", getInstances)
	http.HandleFunc("/api/saveInstance", saveInstance)

	// PROXY
	http.HandleFunc("/api/proxy", getProxies)

	// http://localhost:8090/ will server index.html
	http.Handle("/", http.FileServer(http.Dir("./web")))

	//  nil tells it to use the default router
	http.ListenAndServe(httpPort, nil)

}
