package config

import (
	"encoding/json"
	"errors"
	"os"
)

const TOPIC_PREFIX = "homeassistant"

type Configuration struct {
	Lights    []LightConfiguration `json:"lights"`
	Mqtt      Mqtt                 `json:"mqtt"`
	IpAddress string               `json:"ip_address"`
}

type Mqtt struct {
	IpAddress string `json:"ip_address"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

func LoadConfiguration(filename string) (*Configuration, error) {
	var file *os.File
	var err error
	if file, err = os.Open(filename); err != nil {
		return nil, err
	}

	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := &Configuration{}
	if err := decoder.Decode(configuration); err != nil {
		return nil, err
	}

	// Validate configuration
	if configuration.IpAddress == "" {
		return nil, errors.New("ip_address is required")
	}

	return configuration, nil
}
