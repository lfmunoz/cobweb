package main

import (
	"fmt"
	"os"

	// LOGGING
	"github.com/lfmunoz/cobweb/internal/config"
	log "github.com/sirupsen/logrus"
)

// ________________________________________________________________________________
// callback handlers
// ________________________________________________________________________________

func pwd() {
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	fmt.Println(path)
}

// ________________________________________________________________________________
// MAIN
// ________________________________________________________________________________
func main() {
	pwd()

	local := config.Listener{
		Name:    "http",
		Port:    80,
		Address: "localhost",
	}

	remote := config.Cluster{
		Name:    "http",
		Port:    80,
		Address: "localhost",
	}

	config.BuildListener(local, remote)
}
