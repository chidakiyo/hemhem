package main

import (
	"fmt"
	"github.com/chidakiyo/hemhem/dongle"
	"go.uber.org/zap"
	"os"
	"time"
)

var (
	logger, _ = zap.NewDevelopment()
)

func main() {

	logger.Info("[START]")

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("recover:", err) // panic
		}
	}()

	pwd := os.Getenv("HEMS_PASSWORD")
	rbID := os.Getenv("HEMS_ROUTEB_ID")

	logger.Info(fmt.Sprintf("pass: %s, id:%s", pwd, rbID))

	du := dongle.NewDongleUtil(logger)

	du.Init(pwd, rbID)

	resultCh := make(chan Result, 3)

	go func() {
		for {
			data := <-resultCh
			// ここでbqとかにぶん投げる
			logger.Info(fmt.Sprintf("[OUTPUT] %v", data))
			logger.Info("-----------------------------------------------")
		}
	}()

	t := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-t.C:
			go func() { // たまに止まってしまう時があるので非同期に
			    // TODO タイムアウト処理必要
				du.Fetch(func(time time.Time, watt uint64) {
					resultCh <- Result{
						Time: time,
						Watt: watt,
					}
				})
			}()
		}
	}

}

type Result struct {
	Watt uint64
	Time time.Time
}
