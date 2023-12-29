package main

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/victorjacobs/go-domestia/bridge"
	"github.com/victorjacobs/go-domestia/config"
)

// TODO mark everything as unavailable on shutdown

func main() {
	cfg, err := config.LoadConfiguration("domestia.json")
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	log.Printf("Connecting to %v, managing %v relays", cfg.IpAddress, len(cfg.Lights))

	b, err := bridge.New(cfg)
	if err != nil {
		log.Fatalf("Failed to set up bridge: %v", err)
	}

	for {
		if err := b.Run(); err != nil {
			log.Errorf("Bridge returned: %v, restarting", err)
		} else {
			log.Print("Shutting down")
			break
		}

		time.Sleep(time.Second)
	}
}
