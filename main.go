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

func main() {
	var cfg *config.Configuration
	var err error
	if cfg, err = config.LoadConfiguration("domestia.json"); err != nil {
		log.Printf("Error loading configuration: %v\n", err)
		return
	}

	var domestiaClient *domestia.DomestiaClient
	if domestiaClient, err = domestia.NewDomestiaClient(cfg.IpAddress, cfg.Lights); err != nil {
		log.Printf("Error setting up Domestia client: %v", err)
		return
	}

	mqttOpts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%v:1883", cfg.Mqtt.IpAddress)).SetUsername(cfg.Mqtt.Username).SetPassword(cfg.Mqtt.Password)
	mqttClient := mqtt.NewClient(mqttOpts)
	if t := mqttClient.Connect(); t.Wait() && t.Error() != nil {
		log.Printf("MQTT connection error: %v", t.Error())
		return
	}

	for _, l := range cfg.Lights {
		// Publish configuration for MQTT autodiscovery
		if !l.HiddenInHomeAssistant {
			configTopic := l.ConfigTopic()
			if configJson, err := l.ConfigJson(); err != nil {
				log.Printf("Error marshalling light configuration: %v", err)
				return
			} else if t := mqttClient.Publish(configTopic, 0, true, configJson); t.Wait() && t.Error() != nil {
				log.Printf("MQTT publish failed: %v", t.Error())
				return
			} else {
				log.Printf("MQTT published to %v", configTopic)
			}
		}

		// Subscribe to all light command topics
		cmdTopic := l.CommandTopic()

		log.Printf("MQTT subscribing to %v", cmdTopic)

		if t := mqttClient.Subscribe(cmdTopic, 0, func(client mqtt.Client, msg mqtt.Message) {
			log.Printf("Received: %v", string(msg.Payload()))

			cmd := &LightCommand{}
			if err := json.Unmarshal(msg.Payload(), cmd); err != nil {
				log.Printf("MQTT deserialization failed: %v", err)
				return
			}

			if cmd.State == "ON" {
				domestiaClient.TurnOn(l.Relay)

				// If the command is "on", brightness will never be 0. Therefore if it is 0 here, that means it was absent in the payload
				if cmd.Brightness != 0 {
					domestiaClient.SetBrightness(l.Relay, cmd.Brightness)
				}
			} else {
				domestiaClient.TurnOff(l.Relay)
			}
		}); t.Wait() && t.Error() != nil {
			log.Printf("MQTT receive error: %v", t.Error())
		}
	}

	relayToBrightness := make(map[int]int)

	for {
		if lights, err := domestiaClient.GetState(); err != nil {
			log.Printf("Error: %v", err)
		} else {
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
					if stateJson, err := light.StateJson(); err != nil {
						log.Printf("Error marshalling light state: %v", err)
					} else if t := mqttClient.Publish(configuration.StateTopic(), 0, true, stateJson); t.Wait() && t.Error() != nil {
						log.Printf("MQTT publish error: %v", t.Error())
					} else {
						log.Printf("MQTT published to %v", configuration.StateTopic())
					}
				}
			}
		}

		time.Sleep(2 * time.Second)
	}
}
