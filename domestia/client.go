package domestia

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/victorjacobs/go-domestia/config"
)

const port = 52001

type Client struct {
	mutex              *sync.Mutex
	ipAddress          string
	conn               net.Conn
	lightConfiguration map[int]config.LightConfiguration
}

func NewClient(ipAddress string, lights []config.LightConfiguration) (*Client, error) {
	if ipAddress == "" {
		return nil, errors.New("NewDomestiaClient requires ipAddress")
	}

	lightConfiguration := make(map[int]config.LightConfiguration)

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
	if conn, err := net.Dial("tcp", d.connectUrl()); err != nil {
		return err
	} else if err = conn.SetDeadline(time.Now().Add(time.Second)); err != nil {
		return err
	} else {
		d.conn = conn

		return nil
	}
}

func (d *Client) connectUrl() string {
	return fmt.Sprintf("%v:%v", d.ipAddress, port)
}

func (d *Client) GetState() ([]Light, error) {
	stateCommand := []byte{
		0xff, 0x00, 0x00, 0x01, 0x3c, 0x3c, 0x20,
	}

	var response []byte
	var err error
	if response, err = d.send(stateCommand); err != nil {
		return nil, err
	}

	lights := make([]Light, 0)

	if response[0] != 0xff {
		return lights, nil
	}

	for i, byte := range response[3:] {
		if cfg, ok := d.lightConfiguration[i+1]; ok {
			lights = append(lights, NewLight(cfg, byte))
		}
	}

	return lights, nil
}

func (d *Client) TurnOff(relay int) error {
	return d.toggleRelay(0x0f, relay)
}

func (d *Client) TurnOn(relay int) error {
	return d.toggleRelay(0x0e, relay)
}

func (d *Client) toggleRelay(cmd byte, relay int) error {
	command := []byte{
		0xff, 0x00, 0x00, 0x02, cmd, byte(relay), cmd + byte(relay),
	}

	if response, err := d.send(command); err != nil {
		return err
	} else if string(response) != "OK" {
		return fmt.Errorf("unexpected command response. Expected OK, received %v", string(response))
	}

	return nil
}

func (d *Client) SetBrightness(relay int, brightness int) error {
	brightnessForController := int(float64(brightness) * (64.0 / 255.0))

	command := []byte{
		0xff, 0x00, 0x00, 0x03, 0x10, byte(relay), byte(brightnessForController), byte(0x10 + relay + brightnessForController),
	}

	if response, err := d.send(command); err != nil {
		return err
	} else if string(response) != "OK" {
		return fmt.Errorf("unexpected command response. Expected OK, received %v", string(response))
	}

	return nil
}

func (d *Client) send(command []byte) ([]byte, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if err := d.connect(); err != nil {
		return nil, err
	}
	defer d.conn.Close()

	response := make([]byte, 256)

	if _, err := d.conn.Write(command); err != nil {
		return nil, err
	} else if n, err := d.conn.Read(response); err != nil {
		return nil, err
	} else if n == 0 {
		return nil, errors.New("read 0 bytes")
	}

	return response, nil
}
