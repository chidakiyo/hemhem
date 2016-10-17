package main

import (
	"bufio"
	"fmt"
	"github.com/tarm/serial"
	"strconv"
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
		fmt.Errorf("SKVER がダメ")
	}

	err = d.SKSETPWD(pwd)
	if err != nil {
		// TODO NOOP
		fmt.Errorf("SKSETPWD がダメ")
	}

	err = d.SKSETRBID(rbID)
	if err != nil {
		// TODO NOOP
		fmt.Errorf("SKSETRBID がダメ")
	}

	pan, err := d.SKSCAN()
	fmt.Printf("%#d\n", pan)
	if err != nil {
		// TODO NOOP
		fmt.Errorf("SKSCAN がダメ")
	}

	err = d.SKSREG("S2", pan.Channel)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Set PanID to S3 register...")
	err = d.SKSREG("S3", pan.PanID)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Get IPv6 Addr with SKLL64...")
	ipv6Addr, err := d.SKLL64(pan.Addr)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("IPv6 Addr is " + ipv6Addr)
	fmt.Println("SKJOIN...")
	err = d.SKJOIN(ipv6Addr)
	if err != nil {
		fmt.Println(err)
	}
	b := []byte{0x10, 0x81, 0x00, 0x01, 0x05, 0xFF, 0x01, 0x02, 0x88, 0x01, 0x62, 0x01, 0xE7, 0x00}
	for {
		fmt.Println("SKSENDTO...")
		r, err := d.SKSENDTO("1", ipv6Addr, "0E1A", "1", b)
		if err != nil {
			fmt.Println(err)
		}
		a := strings.Split(r, " ")
		if len(a) != 9 {
			fmt.Println(r)
		}
		if a[7] != "0012" {
			fmt.Println(fmt.Sprintf("%s is not 0012. ", a[7]))
			continue
		}
		o := a[8]
		w, err := strconv.ParseUint(o[len(o)-8:], 16, 0)
		if err != nil {
			fmt.Println(err)
		}
		t := time.Now()
		fmt.Println(t, w)
		data := map[string]string{
			"timestamp": time.Time(t).UTC().Format("2006-01-02T15:04:05.000Z"),
			"watt":      strconv.FormatUint(w, 10),
		}
		fmt.Printf("[OUTPUT] %v", data)
		//err = logger.Post(*tag, data)
		//if err != nil {
		//	log.Error(err)
		//}
		time.Sleep(60 * time.Second)
	}

}

type Dongle struct {
	Baudrate     int
	SerialDevice string
	Port         *serial.Port
}

func NewDongle() *Dongle {
	return &Dongle{
		Baudrate: 115200,
		//SerialDevice: "/dev/ttyUSB0",
		SerialDevice: "/dev/tty.usbserial-A103BTKQ",
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
		fmt.Println(l)
		if strings.Contains(l, "FAIL ") {
			return "", fmt.Errorf("Failed to SKSENDTO. %s", l)
		}
		if strings.Contains(l, "ERXUDP ") {
			return l, nil
		}
	}
	return "", nil
}
