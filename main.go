package main

import (
	"fmt"
	"github.com/hootuu/core/broadcast"
	"github.com/hootuu/core/hotu"
	"github.com/hootuu/core/plugins/linkerx"
	"github.com/hootuu/domain/chain"
	"github.com/hootuu/domain/point"
	"github.com/hootuu/domain/scope"
	"time"
)

func main() {
	//
	//if err := ucs.StartUp(); err != nil {
	//	logger.Logger.Error("ucs.StartUp failed", zap.Error(err))
	//	return
	//}
	//
	//if err := ucs.StartGW(); err != nil {
	//	logger.Logger.Error("ucs.StartGW failed", zap.Error(err))
	//	return
	//}
	//
	//if err := ucs.StartWebui(); err != nil {
	//	logger.Logger.Info("ucs.StartWebui failed", zap.String("err", err.Error()))
	//	return
	//}
	//
	//mq, err := broadcast.NewMQ("vn.join")
	//if err != nil {
	//	logger.Logger.Error("ucs.StartWebui failed", zap.Error(err))
	//	return
	//}
	//mq.RegisterListener(&testListener{})
	//mq.StartListening()

	hotu.Hotu.Init(point.Modes.Angry, scope.HotuAngryLead)
	hotu.Hotu.StartUp()

	go func() {
		first := true
		for {
			data := broadcast.Data{
				VN:        "a",
				Scope:     "b",
				Timestamp: time.Now().UnixMilli(),
				Tag:       []string{"A"},
			}
			if first {
				data.WithData(chain.CreationLink{
					Lead: scope.Lead{
						VN:    "a",
						Scope: "b",
					},
					Code: "order",
				})
			} else {
				data.WithData(chain.Link{
					Lead: scope.Lead{
						VN:    "a",
						Scope: "b",
					},
					Code: "order",
					Data: fmt.Sprintf("ORDERID_%d", time.Now().UnixMilli()),
				})
			}
			err := linkerx.LinkerX.AppendMQ.Publish(data)
			if err != nil {
				fmt.Println(err)
			}
			time.Sleep(10 * time.Second)
		}
	}()
	time.Sleep(10 * time.Hour)
}
