package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// APP CONFIG
type AppConfig struct {
	HttpDir  string
	HttpPort uint
	GrpcPort uint
}

func ReadAppConfig() AppConfig {
	log.Infof("[AppConfig] - reading conf.json...")

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	path := filepath.Join(exPath, "conf.json")
	log.Infof("[AppConfig] - path %s", path)
	file, err := os.Open(path)
	if err != nil {
		log.Infof("[AppConfig] - Not found using defaults")
		return AppConfig{"./web", 8090, 18000}
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	appConfig := AppConfig{}
	err = decoder.Decode(&appConfig)
	if err != nil {
		log.Warnf("[AppConfig] - invalid JSON using defaults")
		return AppConfig{"./web", 8090, 18000}
	}
	log.Infof("[AppConfig] - %+v ", appConfig)
	return appConfig
}
