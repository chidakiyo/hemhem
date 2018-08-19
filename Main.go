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
	executeCh := make(chan string, 3)

	resultHandler :=  func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("recover goroutine. :", err) // panic
			}
		}()
		for {
			data := <-resultCh
			// TODO ここでbqとかにぶん投げる
			logger.Info(fmt.Sprintf("[OUTPUT] %+v", data))
			logger.Info("-----------------------------------------------")
		}
	}

	processor :=  func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("recover goroutine. :", err) // panic
			}
		}()
		for {
			<-executeCh
			// TODO タイムアウト処理必要
			du.Fetch(func(time time.Time, watt uint64) {
				resultCh <- Result{
					Time: time,
					Watt: watt,
				}
			}, executeCh)
		}
	}

	go processor()
	go resultHandler()

	t := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-t.C:
			executeCh <- "" // execute
		}
	}

}

type Result struct {
	Watt uint64
	Time time.Time
}
