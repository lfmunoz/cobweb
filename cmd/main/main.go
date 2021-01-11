package main

import (
	"github.com/lfmunoz/cobweb/internal/config"
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
	appConfig := config.ReadAppConfig()
	go envoy.Start(appConfig)
	web.Start(appConfig)
}
