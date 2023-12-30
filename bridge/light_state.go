package bridge

import (
	"encoding/json"

	"github.com/victorjacobs/go-domestia/domestia"
)

// lightState represents the light state as published over MQTT
type lightState struct {
	State      string `json:"state"`
	Brightness int    `json:"brightness"`
}

func marshalLightToJSON(l *domestia.Light) (string, error) {
	state := &lightState{
		Brightness: brightnessFromDomestia(l.Brightness),
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

// brightnessFromDomestia converts the brightness in domestia.Light to brightness as published over MQTT
func brightnessFromDomestia(brightness uint8) int {
	return int(float32(brightness) * (255.0 / 63.0))
}
