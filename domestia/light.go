package domestia

import (
	"encoding/json"

	"github.com/victorjacobs/go-domestia/config"
)

type Light struct {
	// TODO check what is really required here
	Configuration config.LightConfiguration
	Brightness    int
}

func NewLight(cfg config.LightConfiguration, brightness byte) Light {
	// log.Printf("Raw brightness %v: %v", cfg.Name, brightness)
	return Light{
		Configuration: cfg,
		// TODO something is wrong with this calculation
		Brightness: int(float64(brightness) * (255.0 / 63.0)), // The controller returns brightness [0..63] so convert it to [0..255]
	}
}

type LightState struct {
	State      string `json:"state"`
	Brightness int    `json:"brightness"`
}

func (l *Light) StateJson() (string, error) {
	state := &LightState{
		Brightness: l.Brightness,
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
