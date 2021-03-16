package domestia

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/victorjacobs/go-domestia/config"
)

const PORT = 52001

type DomestiaClient struct {
	mutex              sync.Mutex
	ipAddress          string
	conn               net.Conn
	lightConfiguration map[int]config.LightConfiguration
}

func NewDomestiaClient(ipAddress string, lights []config.LightConfiguration) (*DomestiaClient, error) {
	if ipAddress == "" {
		return nil, errors.New("NewDomestiaClient requires ipAddress")
	}

	lightConfiguration := make(map[int]config.LightConfiguration)

	for _, light := range lights {
		lightConfiguration[light.Relay] = light
	}

	return &DomestiaClient{
		ipAddress:          ipAddress,
		lightConfiguration: lightConfiguration,
	}, nil
}

func (d *DomestiaClient) connect() error {
	if conn, err := net.Dial("tcp", d.connectUrl()); err != nil {
		return err
	} else if err = conn.SetDeadline(time.Now().Add(time.Second)); err != nil {
		return err
	} else {
		d.conn = conn

		return nil
	}
}

func (d *DomestiaClient) connectUrl() string {
	return fmt.Sprintf("%v:%v", d.ipAddress, PORT)
}

func (d *DomestiaClient) GetState() ([]Light, error) {
	stateCommand := []byte{
		0xff, 0x00, 0x00, 0x01, 0x3c, 0x3c, 0x20,
	}

	var response []byte
	var err error
	if response, err = d.send(stateCommand); err != nil {
		return nil, err
	}

	lights := make([]Light, 0)
	for i, byte := range response[3:] {
		if cfg, ok := d.lightConfiguration[i+1]; ok {
			lights = append(lights, NewLight(cfg, byte))
		}
	}

	return lights, nil
}

func (d *DomestiaClient) TurnOff(relay int) error {
	return d.toggleRelay(0x0f, relay)
}

func (d *DomestiaClient) TurnOn(relay int) error {
	return d.toggleRelay(0x0e, relay)
}

func (d *DomestiaClient) toggleRelay(cmd byte, relay int) error {
	command := []byte{
		0xff, 0x00, 0x00, 0x02, cmd, byte(relay), cmd + byte(relay),
	}

	_, err := d.send(command)

	return err
}

func (d *DomestiaClient) SetBrightness(relay int, brightness int) error {
	brightnessForController := int(float64(brightness) * (63.0 / 255.0))

	command := []byte{
		0xff, 0x00, 0x00, 0x03, 0x10, byte(relay), byte(brightnessForController), byte(0x10 + relay + brightnessForController),
	}

	_, err := d.send(command)

	return err
}

func (d *DomestiaClient) send(command []byte) ([]byte, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.connect()
	defer d.conn.Close()

	response := make([]byte, 256)

	log.Printf("Sending %v\n", command)

	var err error
	var n int
	if _, err = d.conn.Write(command); err != nil {
		return nil, err
	} else if n, err = d.conn.Read(response); err != nil {
		return nil, err
	}

	if response[0] != 0xff {
		log.Printf("Received %v\n", string(response))
	} else {
		log.Printf("Received %v bytes", n)
	}

	return response, nil
}
