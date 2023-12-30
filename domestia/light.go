package domestia

import "github.com/victorjacobs/go-domestia/config"

// Light represents a light as retrieved from Domestia controller.
type Light struct {
	Configuration *config.Light
	Brightness    uint8
}

// NewLight creates a new Light struct from given configuration and brightness.
func NewLight(cfg *config.Light, brightness uint8) *Light {
	// If brightness is exactly 1, the relay is not dimmable and on.
	if brightness == 1 {
		brightness = 63
	}

	return &Light{
		Configuration: cfg,
		Brightness:    brightness,
	}
}

func (l *Light) IsMaxBrightness() bool {
	return l.Brightness == 63
}

func (l *Light) IsMinBrightness() bool {
	return l.Brightness == 0
}
