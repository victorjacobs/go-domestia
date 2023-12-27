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
		log.Panicf("Error loading configuration: %v", err)
	}

	log.Printf("Connecting to %v, managing %v relays", cfg.IpAddress, len(cfg.Lights))

	b, err := bridge.New(cfg)
	if err != nil {
		log.Panicf("Failed to set up bridge: %v", err)
	}

	go loopSafely(func() {
		if err := b.PublishLightState(); err != nil {
			panic(err)
		}

		time.Sleep(time.Duration(cfg.RefreshFrequency * 1_000_000))
	})

	select {}
}

func loopSafely(f func()) {
	defer func() {
		if v := recover(); v != nil {
			log.Printf("Panic: %v, restarting", v)
			time.Sleep(time.Second)
			go loopSafely(f)
		}
	}()

	for {
		f()
	}
}
