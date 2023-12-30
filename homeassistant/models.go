package homeassistant

// LightState represents the light state as published over MQTT to Home Assistant.
// Is use for both reporting light state and also the "requested" state through command.
type LightState struct {
	State      string `json:"state"`
	Brightness int    `json:"brightness"`
}
