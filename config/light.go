package config

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Light struct {
	Name     string `json:"name"`      // Light name
	Relay    uint8  `json:"relay"`     // Relay number
	Dimmable bool   `json:"dimmable"`  // Whether the light relay is dimmable
	AlwaysOn bool   `json:"always_on"` // Whether the light relay should always be on, hides the relay in home assistant
}

func (l *Light) EntityId() string {
	return strings.Replace(strings.ToLower(l.Name), " ", "_", -1)
}

func (l *Light) UniqueId() string {
	return fmt.Sprintf("d_%v", l.Relay)
}

func (l *Light) CommandTopic() string {
	return fmt.Sprintf("%v/light/%v/set", topicPrefix, l.EntityId())
}

func (l *Light) StateTopic() string {
	return fmt.Sprintf("%v/light/%v/state", topicPrefix, l.EntityId())
}

func (l *Light) ConfigTopic() string {
	return fmt.Sprintf("homeassistant/light/%v/config", l.EntityId())
}

type LightConfigJson struct {
	Name         string `json:"name"`
	UniqueId     string `json:"unique_id"`
	CommandTopic string `json:"command_topic"`
	StateTopic   string `json:"state_topic"`
	Schema       string `json:"schema"`
	Brightness   bool   `json:"brightness"`
}

func (l *Light) ConfigJson() (string, error) {
	config := LightConfigJson{
		Name:         l.Name,
		UniqueId:     l.UniqueId(),
		CommandTopic: l.CommandTopic(),
		StateTopic:   l.StateTopic(),
		Schema:       "json",
		Brightness:   l.Dimmable,
	}

	if configMarshalled, err := json.Marshal(config); err != nil {
		return "", err
	} else {
		return string(configMarshalled), nil
	}
}
