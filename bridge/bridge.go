package bridge

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/victorjacobs/go-domestia/config"
	"github.com/victorjacobs/go-domestia/domestia"
)

type Bridge struct {
	cfg      *config.Configuration
	domestia *domestia.Client
	// Map to store current brightnesses of lights, used to publish only on changes to state
	relayToBrightness map[int]int
}

type lightCommand struct {
	State      string `json:"state"`
	Brightness int    `json:"brightness"`
}

func New(cfg *config.Configuration) (*Bridge, error) {
	if domestiaClient, err := domestia.NewClient(cfg.IpAddress, cfg.Lights); err != nil {
		return nil, err
	} else {
		return &Bridge{
			cfg:               cfg,
			domestia:          domestiaClient,
			relayToBrightness: make(map[int]int),
		}, nil
	}
}

// Setup function: publishes Home Assistant configuration and subscribes to state updates
func (b *Bridge) SetupLights(mqttClient mqtt.Client) error {
	// Register lights with Homeassistant and subscribe to command topics
	for _, l := range b.cfg.Lights {
		// Bind current l to light
		light := l

		// Publish configuration for MQTT autodiscovery
		if !light.HiddenInHomeAssistant {
			configTopic := light.ConfigTopic()
			if configJson, err := light.ConfigJson(); err != nil {
				return fmt.Errorf("error marshalling light configuration: %v", err)
			} else if t := mqttClient.Publish(configTopic, 0, true, configJson); t.Wait() && t.Error() != nil {
				return fmt.Errorf("MQTT publish failed: %v", t.Error())
			}

			log.Printf("Registered %v with Homeassistant", light.Name)
		}

		// Subscribe to all light command topics
		if t := mqttClient.Subscribe(light.CommandTopic(), 0, func(mqttClient mqtt.Client, msg mqtt.Message) {
			relay := light.Relay
			cmd := &lightCommand{}
			if err := json.Unmarshal(msg.Payload(), cmd); err != nil {
				log.Panicf("MQTT deserialization failed: %v", err)
			}

			if cmd.State == "ON" {
				log.Printf("Turning on %v", light.Name)
				b.domestia.TurnOn(relay)

				if !light.Dimmable {
					b.domestia.SetBrightness(relay, 255)
				} else if cmd.Brightness != 0 {
					b.domestia.SetBrightness(relay, cmd.Brightness)
				}
			} else {
				log.Printf("Turning off %v", light.Name)
				b.domestia.TurnOff(relay)
			}
		}); t.Wait() && t.Error() != nil {
			return fmt.Errorf("MQTT receive error: %v", t.Error())
		}
	}

	return nil
}

// Fetches current state of the controller and publishes updates to mqtt.
// Also makes sure always-on lights are in fact always on. Also makes sure
// that non-dimmable lights are not dimmed.
func (b *Bridge) PublishLightState(mqttClient mqtt.Client) error {
	lights, err := b.domestia.GetState()

	if err != nil {
		return err
	}

	for _, light := range lights {
		configuration := light.Configuration

		var shouldPublishUpdate bool
		if brightness, present := b.relayToBrightness[configuration.Relay]; !present {
			shouldPublishUpdate = true
		} else {
			shouldPublishUpdate = light.Brightness != brightness
		}

		if configuration.AlwaysOn && light.Brightness != 255 {
			// If the light is always-on, and the brightness is not 100%, set it to 100%
			log.Printf("Turning always-on light %v back on", light.Configuration.Name)
			b.domestia.TurnOn(configuration.Relay)
			b.domestia.SetBrightness(configuration.Relay, 255)

			shouldPublishUpdate = false
		} else if !configuration.Dimmable && light.Brightness != 0 && light.Brightness != 255 {
			// If the light is not dimmable and on it should always be set to 100%
			log.Printf("Non-dimmable light %v at brightness %v, resetting", configuration.Name, light.Brightness)

			b.domestia.SetBrightness(configuration.Relay, 255)
			shouldPublishUpdate = false
		} else {
			b.relayToBrightness[configuration.Relay] = light.Brightness
		}

		if shouldPublishUpdate {
			log.Printf("%v changed state", configuration.Name)
			stateTopic := configuration.StateTopic()
			if stateJson, err := light.StateJson(); err != nil {
				return fmt.Errorf("[%v] Error marshalling light state: %v", stateTopic, err)
			} else if t := mqttClient.Publish(stateTopic, 0, true, stateJson); t.Wait() && t.Error() != nil {
				return fmt.Errorf("[%v] Publish error: %v", stateTopic, t.Error())
			}
		}
	}

	return nil
}
