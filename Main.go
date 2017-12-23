package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"os"
	"github.com/chidakiyo/hemhem/dongle"
	"go.uber.org/zap"
)

func main() {

	logger, _ := zap.NewDevelopment()

	logger.Info("[START]")

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("recover:", err) // panic
		}
	}()

	pwd := os.Getenv("HEMS_PASSWORD")
	rbID := os.Getenv("HEMS_ROUTEB_ID")

	logger.Info(fmt.Sprintf("pass: %s, id:%s", pwd, rbID))

	d := dongle.NewDongle()

	logger.Info("Connect...")
	d.Connect()
	logger.Info("Connect OK.")
	defer d.Close()

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
	if err != nil {
		logger.Fatal("IPv6 Address.")
	}

	fmt.Println("IPv6 Addr is " + ipv6Addr)
	fmt.Println("SKJOIN...")
	err = d.SKJOIN(ipv6Addr)
	if err != nil {
		logger.Fatal("SKJOIN.")
	}
	b := []byte{0x10, 0x81, 0x00, 0x01, 0x05, 0xFF, 0x01, 0x02, 0x88, 0x01, 0x62, 0x01, 0xE7, 0x00}
	for {
		logger.Info("SKSENDTO...")
		r, err := d.SKSENDTO("1", ipv6Addr, "0E1A", "1", b)
		if err != nil {
			logger.Fatal("error", zap.Any("err", err))
		}
		a := strings.Split(r, " ")
		if len(a) != 9 {
			logger.Fatal("error", zap.Any("err", err))
		}
		if a[7] != "0012" {
			fmt.Println(fmt.Sprintf("%s is not 0012. ", a[7]))
			continue
		}
		o := a[8]
		w, err := strconv.ParseUint(o[len(o)-8:], 16, 0)
		if err != nil {
			logger.Fatal("error", zap.Any("err", err))
		}
		t := time.Now()
		fmt.Println(t, w)
		data := map[string]string{
			//"timestamp": time.Time(t).UTC().Format("2006-01-02T15:04:05.000Z"),
			"timestamp": time.Time(t).Format("2006-01-02T15:04:05.000Z"),
			"watt":      strconv.FormatUint(w, 10),
		}
		logger.Info(fmt.Sprintf("[OUTPUT] %v", data))
		logger.Info("-----------------------------------------------")
		//err = logger.Post(*tag, data)
		//if err != nil {
		//	log.Error(err)
		//}
		time.Sleep(5 * time.Second)
		//time.Sleep(60 * time.Second)
	}

}

