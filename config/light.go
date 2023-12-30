package config

import (
	"encoding/json"
	"fmt"

	"github.com/victorjacobs/go-domestia/homeassistant"
)

type Light struct {
	Name     string `json:"name"`      // Light name
	Relay    uint8  `json:"relay"`     // Relay number
	Dimmable bool   `json:"dimmable"`  // Whether the light relay is dimmable
	AlwaysOn bool   `json:"always_on"` // Whether the light relay should always be on, hides the relay in home assistant
}

func (l *Light) HomeAssistant() *homeassistant.LightConfiguration {
	return homeassistant.NewLightConfiguration(l.Name, fmt.Sprintf("d_%v", l.Relay), l.Dimmable)
}

func (l *Light) HomeAssistantRegistrationJSON() (string, error) {
	config := l.HomeAssistant()

	if configMarshalled, err := json.Marshal(config); err != nil {
		return "", err
	} else {
		return string(configMarshalled), nil
	}
}
