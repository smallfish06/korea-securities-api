package main

import (
	"flag"
	"log"

	"github.com/smallfish06/kr-broker-api/internal/config"
	"github.com/smallfish06/kr-broker-api/internal/server"
)

func main() {
	configFile := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	srv := server.New(cfg)
	if err := srv.Run(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
