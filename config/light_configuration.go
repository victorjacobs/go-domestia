package config

import (
	"encoding/json"
	"fmt"
	"strings"
)

type LightConfiguration struct {
	Name                  string `json:"name"`
	Relay                 int    `json:"relay"`
	Dimmable              bool   `json:"dimmable"`
	AlwaysOn              bool   `json:"always_on"`
	HiddenInHomeAssistant bool   `json:"hidden_in_home_assistant"`
}

func (l *LightConfiguration) EntityId() string {
	return fmt.Sprintf("d_%v", strings.Replace(strings.ToLower(l.Name), " ", "_", -1))
}

func (l *LightConfiguration) UniqueId() string {
	return fmt.Sprintf("d_%v", l.Relay)
}

func (l *LightConfiguration) CommandTopic() string {
	return fmt.Sprintf("%v/light/%v/set", TOPIC_PREFIX, l.EntityId())
}

func (l *LightConfiguration) StateTopic() string {
	return fmt.Sprintf("%v/light/%v/state", TOPIC_PREFIX, l.EntityId())
}

func (l *LightConfiguration) ConfigTopic() string {
	return fmt.Sprintf("%v/light/%v/config", TOPIC_PREFIX, l.EntityId())
}

type LightConfigJson struct {
	Name         string `json:"name"`
	UniqueId     string `json:"unique_id"`
	CommandTopic string `json:"command_topic"`
	StateTopic   string `json:"state_topic"`
	Schema       string `json:"schema"`
	Brightness   bool   `json:"brightness"`
}

func (l *LightConfiguration) ConfigJson() (string, error) {
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
