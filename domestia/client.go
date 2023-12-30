package domestia

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/victorjacobs/go-domestia/config"
)

type relayState int

const (
	relayStateOff relayState = iota
	relayStateOn
)

type Client struct {
	mutex              *sync.Mutex
	ipAddress          string
	conn               net.Conn
	lightConfiguration map[uint8]config.Light
}

func NewClient(ipAddress string, lights []config.Light) (*Client, error) {
	if ipAddress == "" {
		return nil, errors.New("NewDomestiaClient requires ipAddress")
	}

	lightConfiguration := make(map[uint8]config.Light)

	for _, light := range lights {
		lightConfiguration[light.Relay] = light
	}

	return &Client{
		mutex:              new(sync.Mutex),
		ipAddress:          ipAddress,
		lightConfiguration: lightConfiguration,
	}, nil
}

func (d *Client) connect() error {
	connectURL := fmt.Sprintf("%v:%v", d.ipAddress, 52001)

	if conn, err := net.Dial("tcp", connectURL); err != nil {
		return err
	} else if err = conn.SetDeadline(time.Now().Add(time.Second)); err != nil {
		return err
	} else {
		d.conn = conn

		return nil
	}
}

// GetState queries the controller and returns the current state of all lights.
func (d *Client) GetState() ([]Light, error) {
	response, err := d.send([]byte{0x3c})
	if err != nil {
		return nil, err
	}

	var lights []Light

	if response[0] != 0xff {
		return lights, nil
	}

	for i, relayByte := range response[3:] {
		if cfg, ok := d.lightConfiguration[uint8(i+1)]; ok {
			lights = append(lights, NewLight(cfg, relayByte))
		}
	}

	return lights, nil
}

// TurnOff turns off relay with given index.
func (d *Client) TurnOff(relay uint8) error {
	return d.setRelayState(relay, relayStateOff)
}

// TurnOn turns on relay with given index.
func (d *Client) TurnOn(relay uint8) error {
	return d.setRelayState(relay, relayStateOn)
}

func (d *Client) setRelayState(relay uint8, state relayState) error {
	var toggleCommand byte
	switch state {
	case relayStateOn:
		toggleCommand = 0x0e
	case relayStateOff:
		toggleCommand = 0x0f
	}

	command := []byte{toggleCommand, relay}

	if response, err := d.send(command); err != nil {
		return err
	} else if string(response) != "OK" {
		return fmt.Errorf("unexpected command response. Expected OK, received %v", string(response))
	}

	return nil
}

// SetBrightness sets brightness of relay to given brightness. Brightness is an uint8, so values between 0 and 63.
func (d *Client) SetBrightness(relay uint8, brightness uint8) error {
	command := []byte{
		0x10, relay, brightness,
	}

	if response, err := d.send(command); err != nil {
		return err
	} else if string(response) != "OK" {
		return fmt.Errorf("unexpected command response. Expected OK, received %v", string(response))
	}

	return nil
}

// SetMaxBrightness sets given relay index to maximum brightness.
func (d *Client) SetMaxBrightness(relay uint8) error {
	return d.SetBrightness(relay, 63)
}

func (d *Client) send(command []byte) ([]byte, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if err := d.connect(); err != nil {
		return nil, err
	}
	defer d.conn.Close()

	packed := packCommand(command)

	response := make([]byte, 256)

	if _, err := d.conn.Write(packed); err != nil {
		return nil, err
	} else if n, err := d.conn.Read(response); err != nil {
		return nil, err
	} else if n == 0 {
		return nil, errors.New("read 0 bytes")
	}

	return response, nil
}

// packCommand packs a command into a controller message
func packCommand(cmd []byte) []byte {
	var buf bytes.Buffer

	buf.Write([]byte{0xff, 0x00, 0x00})
	buf.WriteByte(byte(len(cmd)))
	buf.Write(cmd)

	var checksum uint8
	for _, c := range cmd {
		checksum += c
	}

	buf.WriteByte(checksum)

	return buf.Bytes()
}
