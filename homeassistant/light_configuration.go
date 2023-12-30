package homeassistant

import "github.com/victorjacobs/go-domestia/config"

// LightConfiguration represents a Home Assistant light, as used during light registration
type LightConfiguration struct {
	Name         string `json:"name"`
	UniqueId     string `json:"unique_id"`
	CommandTopic string `json:"command_topic"`
	StateTopic   string `json:"state_topic"`
	Schema       string `json:"schema"`
	Brightness   bool   `json:"brightness"`

	AlwaysOn bool
}

func NewLightConfiguration(l *config.Light) *LightConfiguration {
	return nil
}
