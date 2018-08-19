package dongle

import (
	"bufio"
	"fmt"
	"github.com/tarm/serial"
	"runtime"
	"strings"
)

func NewDongle() *Dongle {
	d := &Dongle{
		Baudrate: 115200,
	}

	switch runtime.GOOS {
	case "darwin":
		// mac
		d.SerialDevice = "/dev/tty.usbserial-A103BTKQ"
	default:
		// raspberry pi.
		d.SerialDevice = "/dev/ttyUSB0"
	}

	return d
}

type Dongle struct {
	Baudrate     int
	SerialDevice string
	Port         *serial.Port
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

func (b *Dongle) SKSREG(k, v string) error {
	err := b.write("SKSREG " + k + " " + v + "\r\n")
	if err != nil {
		return err
	}
	_, err = b.readUntilOK()
	if err != nil {
		return err
	}
	return nil
}

func (b *Dongle) SKLL64(addr string) (string, error) {
	err := b.write("SKLL64 " + addr + "\r\n")
	if err != nil {
		return "", err
	}
	reader := bufio.NewReader(b.Port)
	r, _, err := reader.ReadLine()
	if err != nil {
		return "", err
	}
	fmt.Println(r)
	r, _, err = reader.ReadLine()
	if err != nil {
		return "", err
	}
	fmt.Println(r)
	return string(r), nil
}

func (b *Dongle) SKJOIN(ipv6Addr string) error {
	err := b.write("SKJOIN " + ipv6Addr + "\r\n")
	if err != nil {
		return err
	}
	reader := bufio.NewReader(b.Port)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		l := scanner.Text()
		fmt.Print(l)
		if strings.Contains(l, "FAIL ") {
			return fmt.Errorf("Failed to SKJOIN. %s", l)
		}
		if strings.Contains(l, "EVENT 25 ") {
			break
		}
	}
	if scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	return nil
}

func (b *Dongle) SKSENDTO(handle, ipAddr, port, sec string, data []byte) (string, error) {
	s := fmt.Sprintf("SKSENDTO %s %s %s %s %.4X ", handle, ipAddr, port, sec, len(data))
	d := append([]byte(s), data[:]...)
	d = append(d, []byte("\r\n")[:]...)
	_, err := b.Port.Write(d)
	if err != nil {
		return "", err
	}
	reader := bufio.NewReader(b.Port)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		l := scanner.Text()
		if l != "" {
			fmt.Println("[RESPONSE] >> " + l)
		}
		if strings.Contains(l, "FAIL ") {
			return "", fmt.Errorf("Failed to SKSENDTO. %s", l)
		}
		if strings.Contains(l, "ERXUDP ") {
			return l, nil
		}
	}
	return "", nil
}
