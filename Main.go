package main

import (
	"bufio"
	"fmt"
	"github.com/tarm/serial"
	"strings"
	"time"
)

func main() {

	pwd := ""
	rbID := ""

	fmt.Println("Hello go")

	d := NewDongle()

	d.Connect()
	fmt.Println("Open")
	defer d.Close()

	time.Sleep(time.Second * 1)

	v, err := d.SKVER()
	fmt.Printf("%#d\n", v)
	if err != nil {
		// TODO NOOP
	}

	err = d.SKSETPWD(pwd)
	if err != nil {
		// TODO NOOP
	}

	err = d.SKSETRBID(rbID)
	if err != nil {
		// TODO NOOP
	}

	pan, err := d.SKSCAN()
	fmt.Printf("%#d\n", pan)
	if err != nil {
		// TODO NOOP
	}

}

type Dongle struct {
	Baudrate     int
	SerialDevice string
	Port         *serial.Port
}

func NewDongle() *Dongle {
	return &Dongle{
		Baudrate:     115200,
		SerialDevice: "/dev/ttyUSB0",
	}
}

func (b *Dongle) Connect() error {
	c := &serial.Config{
		Name: b.SerialDevice,
		Baud: b.Baudrate,
	}
	s, err := serial.OpenPort(c)
	if err != nil {
		return err
	}
	b.Port = s
	return nil
}

func (b *Dongle) Close() {
	b.Port.Close()
}

func (b *Dongle) SKVER() (string, error) {
	err := b.write("SKVER\r\n")
	if err != nil {
		return "", err
	}
	lines, err := b.readUntilOK()
	if err != nil {
		return "", err
	}
	return strings.Split(lines[1], " ")[1], nil
}

func (b *Dongle) write(s string) error {
	_, err := b.Port.Write([]byte(s))
	if err != nil {
		return err
	}
	return nil
}

func (b *Dongle) readUntilOK() ([]string, error) {
	reader := bufio.NewReader(b.Port)
	scanner := bufio.NewScanner(reader)
	var reply []string
	for scanner.Scan() {
		l := scanner.Text()
		reply = append(reply, l)
		if l == "OK" {
			break
		}
	}
	return reply, nil
}

func (b *Dongle) SKSETPWD(pwd string) error {
	err := b.write("SKSETPWD C " + pwd + "\r\n")
	if err != nil {
		return err
	}
	return nil

}

func (b *Dongle) SKSETRBID(rbid string) error {
	err := b.write("SKSETRBID " + rbid + "\r\n")
	if err != nil {
		return err
	}
	return nil
}

type PAN struct {
	Channel     string
	ChannelPage string
	PanID       string
	Addr        string
	LQI         string
	PairID      string
}

func (b *Dongle) SKSCAN() (*PAN, error) {
	err := b.write("SKSCAN 2 FFFFFFFF 6\r\n")
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(b.Port)
	scanner := bufio.NewScanner(reader)
	pan := &PAN{}
	for scanner.Scan() {
		l := scanner.Text()
		switch {
		case strings.Contains(l, "Channel:"):
			pan.Channel = strings.Split(l, ":")[1]
		case strings.Contains(l, "Channel Page:"):
			pan.ChannelPage = strings.Split(l, ":")[1]
		case strings.Contains(l, "Pan ID:"):
			pan.PanID = strings.Split(l, ":")[1]
		case strings.Contains(l, "Addr:"):
			pan.Addr = strings.Split(l, ":")[1]
		case strings.Contains(l, "LQI:"):
			pan.LQI = strings.Split(l, ":")[1]
		case strings.Contains(l, "PairID:"):
			pan.PairID = strings.Split(l, ":")[1]
		}
		if strings.Contains(l, "EVENT 22 ") {
			break
		}
	}
	return pan, nil
}
