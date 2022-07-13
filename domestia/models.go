package domestia

import (
	"encoding/json"

	"github.com/victorjacobs/go-domestia/config"
)

type Light struct {
	Configuration config.LightConfiguration
	Brightness    int
}

func NewLight(cfg config.LightConfiguration, brightness byte) Light {
	brightnessFloat := float64(brightness)
	// If brightness is exactly 1, the relay is not dimmable and on.
	if brightnessFloat == 1.0 {
		brightnessFloat = 63
	}

	return Light{
		Configuration: cfg,
		Brightness:    int(brightnessFloat * (255.0 / 63.0)), // The controller returns brightness [0..63] so convert it to [0..255]
	}
}

type LightJson struct {
	State      string `json:"state"`
	Brightness int    `json:"brightness"`
}

func (l *Light) StateJson() (string, error) {
	state := &LightJson{
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
