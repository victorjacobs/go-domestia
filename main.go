package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/victorjacobs/go-domestia/config"
	"github.com/victorjacobs/go-domestia/domestia"
)

type LightCommand struct {
	State      string `json:"state"`
	Brightness int    `json:"brightness"`
}

// TODO mark everything as unavailable on shutdown

func main() {
	var cfg *config.Configuration
	var err error
	if cfg, err = config.LoadConfiguration("domestia.json"); err != nil {
		log.Panicf("Error loading configuration: %v", err)
	}

	log.Printf("Connecting to %v, managing %v relays", cfg.IpAddress, len(cfg.Lights))

	var domestiaClient *domestia.DomestiaClient
	if domestiaClient, err = domestia.NewDomestiaClient(cfg.IpAddress, cfg.Lights); err != nil {
		log.Panicf("Error setting up Domestia client: %v", err)
	}

	mqttClient := mqtt.NewClient(cfg.Mqtt.ClientOptions())
	if t := mqttClient.Connect(); t.Wait() && t.Error() != nil {
		log.Panicf("MQTT connection error: %v", t.Error())
	}

	// Register lights with Homeassistant and subscribe to command topics
	for _, l := range cfg.Lights {
		// Bind current l to light
		light := l

		// Publish configuration for MQTT autodiscovery
		if !light.HiddenInHomeAssistant {
			configTopic := light.ConfigTopic()
			if configJson, err := light.ConfigJson(); err != nil {
				log.Panicf("Error marshalling light configuration: %v", err)
			} else if t := mqttClient.Publish(configTopic, 0, true, configJson); t.Wait() && t.Error() != nil {
				log.Panicf("MQTT publish failed: %v", t.Error())
			}

			log.Printf("Registered %v with Homeassistant", light.Name)
		}

		// Subscribe to all light command topics
		if t := mqttClient.Subscribe(light.CommandTopic(), 0, func(client mqtt.Client, msg mqtt.Message) {
			relay := light.Relay
			cmd := &LightCommand{}
			if err := json.Unmarshal(msg.Payload(), cmd); err != nil {
				log.Panicf("MQTT deserialization failed: %v", err)
			}

			if cmd.State == "ON" {
				log.Printf("Turning on %v", light.Name)
				domestiaClient.TurnOn(relay)

				// If the command is "on", brightness will never be 0. Therefore if it is 0 here, that means it was missing in the payload
				if cmd.Brightness != 0 {
					domestiaClient.SetBrightness(relay, cmd.Brightness)
				}
			} else {
				log.Printf("Turning off %v", light.Name)
				domestiaClient.TurnOff(relay)
			}
		}); t.Wait() && t.Error() != nil {
			log.Printf("MQTT receive error: %v", t.Error())
		}
	}

	// Map to store current brightnesses of lights, used to publish only on changes to state
	relayToBrightness := make(map[int]int)

	go loopSafely(func() {
		lights, err := domestiaClient.GetState()

		if err != nil {
			panic(err)
		}

		for _, light := range lights {
			configuration := light.Configuration

			var shouldPublishUpdate bool
			if brightness, present := relayToBrightness[configuration.Relay]; !present {
				shouldPublishUpdate = true
			} else {
				shouldPublishUpdate = light.Brightness != brightness
			}

			if light.Brightness == 0 && configuration.AlwaysOn {
				domestiaClient.TurnOn(configuration.Relay)
				shouldPublishUpdate = false
			} else {
				relayToBrightness[configuration.Relay] = light.Brightness
			}

			if shouldPublishUpdate {
				log.Printf("%v changed state", configuration.Name)
				stateTopic := configuration.StateTopic()
				if stateJson, err := light.StateJson(); err != nil {
					panic(fmt.Sprintf("[%v] Error marshalling light state: %v", stateTopic, err))
				} else if t := mqttClient.Publish(stateTopic, 0, true, stateJson); t.Wait() && t.Error() != nil {
					panic(fmt.Sprintf("[%v] Publish error: %v", stateTopic, t.Error()))
				}
			}
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
