package dongle

import (
	"time"
	"fmt"
	"go.uber.org/zap"
	"strings"
	"strconv"
)

func NewDongleUtil(l *zap.Logger) *DongleUtil {
	return &DongleUtil{
		Logger:l,
	}
}

type DongleUtil struct {
	Logger *zap.Logger
	Dongle *Dongle
	Ipv6addr string
}

func (du *DongleUtil) Init(pwd string, rbID string) (*Dongle, error) {

	d := NewDongle()
	du.Dongle = d // TODO
	logger := du.Logger // TODO

	logger.Info("Connect...")
	d.Connect()
	logger.Info("Connect OK.")
	//defer d.Close()

	logger.Info("Wait 1sec...")
	time.Sleep(time.Second * 1)
	logger.Info("Wait complete.")

	logger.Info("SKVER...")
	v, err := d.SKVER()
	logger.Info(fmt.Sprintf("SKVER Response : %s", v))
	if err != nil {
		logger.Fatal("SKVER.")
	}
	logger.Info("SKVER OK.")

	err = d.SKSETPWD(pwd)
	if err != nil {
		logger.Fatal("SKSETPWD.")
	}

	err = d.SKSETRBID(rbID)
	if err != nil {
		logger.Fatal("SKSETRBID.")
	}

	pan, err := d.SKSCAN()
	fmt.Printf("%#v\n", pan)
	if err != nil {
		logger.Fatal("SKSCAN.")
	}

	err = d.SKSREG("S2", pan.Channel)
	if err != nil {
		logger.Fatal("SKSREG S2.")
	}

	fmt.Println("Set PanID to S3 register...")
	err = d.SKSREG("S3", pan.PanID)
	if err != nil {
		logger.Fatal("SKSREG S3.")
	}
	fmt.Println("Get IPv6 Addr with SKLL64...")
	ipv6Addr, err := d.SKLL64(pan.Addr)
	du.Ipv6addr = ipv6Addr // TODO
	if err != nil {
		logger.Fatal("IPv6 Address.")
	}

	fmt.Println("IPv6 Addr is " + ipv6Addr)
	fmt.Println("SKJOIN...")
	err = d.SKJOIN(ipv6Addr)
	if err != nil {
		logger.Fatal("SKJOIN.")
	}

	return d, nil
}

var b = []byte{0x10, 0x81, 0x00, 0x01, 0x05, 0xFF, 0x01, 0x02, 0x88, 0x01, 0x62, 0x01, 0xE7, 0x00}

func (du *DongleUtil) Fetch(f func(time time.Time, watt uint64)) {

	logger := du.Logger // TODO

	logger.Info("SKSENDTO...")
	r, err := du.Dongle.SKSENDTO("1", du.Ipv6addr, "0E1A", "1", b)
	if err != nil {
		logger.Fatal("error", zap.Any("err", err))
	}
	a := strings.Split(r, " ")
	if len(a) != 9 {
		logger.Fatal("error", zap.Any("err", err))
	}
	if a[7] != "0012" {
		fmt.Println(fmt.Sprintf("%s is not 0012. ", a[7]))
		//continue // TODO 再実行
	}
	o := a[8]
	w, err := strconv.ParseUint(o[len(o)-8:], 16, 0)
	if err != nil {
		logger.Fatal("error", zap.Any("err", err))
	}
	t := time.Now()
	logger.Info(fmt.Sprintf("%+v", t) + " : " +  fmt.Sprintf("%d", w))

	f(t, w) // output
}
