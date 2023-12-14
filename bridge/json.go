package bridge

import (
	"encoding/json"

	"github.com/victorjacobs/go-domestia/domestia"
)

type lightJSON struct {
	State      string `json:"state"`
	Brightness int    `json:"brightness"`
}

// TODO potentially change everything to be pointers? tbd
func marshalLightToJSON(l domestia.Light) (string, error) {
	state := &lightJSON{
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
