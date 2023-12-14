package bridge

// lightCommand represents a command coming in over MQTT
type lightCommand struct {
	State      string `json:"state"`
	Brightness int    `json:"brightness"`
}

// BrightnessForDomestia returns brightness converted to Domestia controller
func (l *lightCommand) BrightnessForDomestia() uint8 {
	return uint8(float32(l.Brightness) * (63.0 / 255.0))
}
