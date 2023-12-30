package homeassistant

import (
	"fmt"
	"strings"
)

// LightConfiguration represents a Home Assistant light, as used during light registration
type LightConfiguration struct {
	ConfigTopic string

	Name         string `json:"name"`
	UniqueId     string `json:"unique_id"`
	CommandTopic string `json:"command_topic"`
	StateTopic   string `json:"state_topic"`
	Schema       string `json:"schema"`
	Brightness   bool   `json:"brightness"`
}

func NewLightConfiguration(name string, uniqueId string, dimmable bool) *LightConfiguration {
	entityId := strings.Replace(strings.ToLower(name), " ", "_", -1)

	return &LightConfiguration{
		ConfigTopic:  fmt.Sprintf("homeassistant/light/%v/config", entityId),
		Name:         name,
		UniqueId:     uniqueId,
		CommandTopic: fmt.Sprintf("domestia/light/%v/set", entityId),
		StateTopic:   fmt.Sprintf("domestia/light/%v/state", entityId),
		Schema:       "json",
		Brightness:   dimmable,
	}
}
