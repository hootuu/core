package main

import (
	"fmt"
	"github.com/hootuu/utils/errors"
	"github.com/hootuu/utils/logger"
	"github/hootuu/core/broadcast"
	"github/hootuu/core/ucs"
	"go.uber.org/zap"
	"time"
)

type testListener struct {
}

func (t testListener) GetCode() string {
	return "testListener"
}

func (t testListener) Care(msg *broadcast.Message) bool {
	return true
}

func (t testListener) Deal(msg *broadcast.Message) *errors.Error {
	//fmt.Println("deal msg:::::", msg.Summary())
	return nil
}

func main() {

	if err := ucs.StartUp(); err != nil {
		logger.Logger.Error("ucs.StartUp failed", zap.Error(err))
		return
	}

	if err := ucs.StartGW(); err != nil {
		logger.Logger.Error("ucs.StartGW failed", zap.Error(err))
		return
	}

	if err := ucs.StartWebui(); err != nil {
		logger.Logger.Info("ucs.StartWebui failed", zap.String("err", err.Error()))
		return
	}

	mq, err := broadcast.NewMQ("vn.join")
	if err != nil {
		logger.Logger.Error("ucs.StartWebui failed", zap.Error(err))
		return
	}
	mq.RegisterListener(&testListener{})
	mq.StartListening()
	go func() {
		for {
			err := mq.Publish(broadcast.Data{
				VN:        "a",
				Scope:     "b",
				ReplyID:   "c",
				Data:      "d",
				Timestamp: time.Now().UnixMilli(),
				Tag:       []string{"A"},
			})
			if err != nil {
				fmt.Println(err)
			}
			time.Sleep(10 * time.Second)
		}
	}()
	time.Sleep(10 * time.Hour)
}
