package main

import (
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
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

	opts := cfg.Mqtt.ClientOptions()
	// Configure MQTT subscriptions in the ConnectHandler to make sure they are set up after reconnect
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		if err := b.SetupLights(client); err != nil {
			log.Panicf("Failed to register with MQTT: %v", err)
		}
	})

	mqttClient := mqtt.NewClient(opts)
	if t := mqttClient.Connect(); t.Wait() && t.Error() != nil {
		log.Panicf("MQTT connection error: %v", t.Error())
	}

	go loopSafely(func() {
		if err := b.PublishLightState(mqttClient); err != nil {
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
