package bridge

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/victorjacobs/go-domestia/config"
	"github.com/victorjacobs/go-domestia/domestia"
)

type Bridge struct {
	cfg      *config.Configuration
	domestia *domestia.Client
	// Channel to trigger an pull and publish state from controller
	updateChannel chan bool
	// Map to store current brightnesses of lights, used to publish only on changes to state
	relayToBrightness map[uint8]uint8
}

func New(cfg *config.Configuration) (*Bridge, error) {
	domestiaClient, err := domestia.NewClient(cfg.IpAddress, cfg.Lights)
	if err != nil {
		return nil, err
	}

	return &Bridge{
		cfg:               cfg,
		domestia:          domestiaClient,
		relayToBrightness: make(map[uint8]uint8),
		updateChannel:     make(chan bool),
	}, nil
}

// Run runs the bridge, blocking. If this function returns an error it can be restarted.
// If it returns nil, it was cleanly shut down.
func (b *Bridge) Run() error {
	mqttClient, err := b.connectMQTT()
	if err != nil {
		return err
	}

	defer func() {
		mqttClient.Disconnect(100)
	}()

	ticker := time.NewTicker(time.Duration(b.cfg.RefreshFrequency) * time.Millisecond)

	for {
		select {
		case <-ticker.C:
		case <-b.updateChannel:
		}

		if err := b.publishLightState(mqttClient); err != nil {
			return err
		}
	}
}

// connectMQTT creates and connects MQTT client
func (b *Bridge) connectMQTT() (mqtt.Client, error) {
	opts := b.cfg.Mqtt.ClientOptions()
	// Configure MQTT subscriptions in the ConnectHandler to make sure they are set up after reconnect
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		if err := b.setupLights(client); err != nil {
			log.Panicf("Failed to register with MQTT: %v", err)
		}
	})

	mqttClient := mqtt.NewClient(opts)
	if t := mqttClient.Connect(); t.Wait() && t.Error() != nil {
		return nil, fmt.Errorf("MQTT connection error: %w", t.Error())
	}

	return mqttClient, nil
}

// setupLights publishes Home Assistant configuration and subscribes to state updates
func (b *Bridge) setupLights(mqttClient mqtt.Client) error {
	for _, l := range b.cfg.Lights {
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
					b.domestia.SetMaxBrightness(relay)
				} else if cmd.Brightness != 0 {
					b.domestia.SetBrightness(relay, cmd.BrightnessForDomestia())
				}
			} else {
				log.Printf("Turning off %v", light.Name)
				b.domestia.TurnOff(relay)
			}

			// Trigger pulling and publishing controller state
			b.updateChannel <- true
		}); t.Wait() && t.Error() != nil {
			return fmt.Errorf("MQTT receive error: %v", t.Error())
		}
	}

	return nil
}

// Fetches current state of the controller and publishes updates to mqtt.
// Also makes sure always-on lights are in fact always on. Also makes sure
// that non-dimmable lights are not dimmed.
func (b *Bridge) publishLightState(mqttClient mqtt.Client) error {
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

		if configuration.AlwaysOn && !light.IsMaxBrightness() {
			// If the light is always-on, and the brightness is not 100%, set it to 100%
			log.Printf("Turning always-on light %v back on", light.Configuration.Name)

			b.domestia.TurnOn(configuration.Relay)
			b.domestia.SetMaxBrightness(configuration.Relay)

			shouldPublishUpdate = false
		} else if !configuration.Dimmable && !light.IsMinBrightness() && !light.IsMaxBrightness() {
			// If the light is not dimmable and on it should always be set to 100%
			log.Printf("Non-dimmable light %v at brightness %v, resetting", configuration.Name, light.Brightness)

			b.domestia.SetMaxBrightness(configuration.Relay)

			shouldPublishUpdate = false
		} else {
			b.relayToBrightness[configuration.Relay] = light.Brightness
		}

		if shouldPublishUpdate {
			log.Printf("%v changed state", configuration.Name)

			stateTopic := configuration.StateTopic()
			if stateJson, err := marshalLightToJSON(light); err != nil {
				return fmt.Errorf("[%v] Error marshalling light state: %v", stateTopic, err)
			} else if t := mqttClient.Publish(stateTopic, 0, true, stateJson); t.Wait() && t.Error() != nil {
				return fmt.Errorf("[%v] Publish error: %v", stateTopic, t.Error())
			}
		}
	}

	return nil
}
