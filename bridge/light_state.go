package bridge

import (
	"encoding/json"

	"github.com/victorjacobs/go-domestia/domestia"
	"github.com/victorjacobs/go-domestia/homeassistant"
)

func homeAssistantStateJSON(l *domestia.Light) (string, error) {
	state := &homeassistant.LightState{
		Brightness: homeAssistantBrightness(l),
	}

	if l.Brightness != 0 {
		state.State = "ON"
	} else {
		state.State = "OFF"
	}

	if stateMarshalled, err := json.Marshal(state); err != nil {
		return "", err
	} else {
		return string(stateMarshalled), nil
	}
}

// homeAssistantBrightness converts the brightness in domestia.Light to brightness as published over MQTT
func homeAssistantBrightness(l *domestia.Light) int {
	return int(float32(l.Brightness) * (255.0 / 63.0))
}

// domestiaBrightness returns brightness converted to Domestia controller
func domestiaBrightness(l *homeassistant.LightState) uint8 {
	return uint8(float32(l.Brightness) * (63.0 / 255.0))
}
