package homeassistant

// LightState represents light state on Home Assistant. Either as current state in
// the state update topic, or requested state in command topic.
type LightState struct {
	State      string `json:"state"`
	Brightness int    `json:"brightness"`
}
