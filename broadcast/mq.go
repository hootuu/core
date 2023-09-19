package broadcast

import (
	"context"
	"fmt"
	"github.com/hootuu/core/hotu/here"
	"github.com/hootuu/utils/errors"
	"github.com/hootuu/utils/sys"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"go.uber.org/zap"
	"strings"
	"sync"
)

const (
	rootProtocol      = "HOTU"
	defaultBufferSize = 256
)

type MQ struct {
	Topic         string
	listenerArray []Listener
	lock          sync.Mutex
	topic         *pubsub.Topic
	subscription  *pubsub.Subscription
}

func NewMQ(topic string) (*MQ, *errors.Error) {
	mq := &MQ{
		Topic:         strings.ToUpper(topic),
		listenerArray: []Listener{},
	}
	err := mq.doInit()
	if err != nil {
		return nil, err
	}
	return mq, nil
}

func (mq *MQ) RegisterListener(listener Listener) {
	mq.lock.Lock()
	defer mq.lock.Unlock()
	for _, lis := range mq.listenerArray {
		if lis.GetCode() == listener.GetCode() {
			gLogger.Error("A duplicate listener", zap.String("code", listener.GetCode()))
			return
		}
	}
	mq.listenerArray = append(mq.listenerArray, listener)
}

func (mq *MQ) Publish(msgData Data) *errors.Error {
	data, err := msgData.ToBytes()
	if err != nil {
		return err
	}
	nErr := mq.topic.Publish(context.Background(), data)
	if nErr != nil {
		gLogger.Error("mq.topic.Publish failed", zap.Error(nErr), zap.Any("data", msgData))
		return errors.Sys("mq.Publish failed", nErr)
	}
	if sys.RunMode.IsRd() {
		gLogger.Info("mq.topic.Publish", zap.Int64("data", msgData.Timestamp))
	}
	return nil
}

func (mq *MQ) StartListening() {
	go func() {
		for {
			payload, nErr := mq.subscription.Next(context.Background())
			if nErr != nil {
				gLogger.Error("subscription.Next failed", zap.Error(nErr))
				continue
			}
			if here.Here.ID() == payload.GetFrom().String() {
				continue
			}
			if sys.RunMode.IsRd() {
				gLogger.Info("Get Message", zap.String("message", payload.GetTopic()))
			}
			msg, err := MessageOf(payload)
			if err != nil {
				gLogger.Error("payload invalid", zap.Error(err))
				continue
			}
			for _, listener := range mq.listenerArray {
				ctx := context.Background()
				care := listener.Care(ctx, msg)
				if !care {
					continue
				}
				if sys.RunMode.IsRd() {
					gLogger.Info("deal message",
						zap.String("listener", listener.GetCode()),
						zap.String("message", msg.Summary()))
				}
				err = listener.Deal(ctx, msg)
				if err != nil {
					gLogger.Error("deal message failed",
						zap.String("listener", listener.GetCode()),
						zap.String("message", msg.Summary()),
						zap.Error(err))
				}
			}
		}
	}()
}

func (mq *MQ) doInit() *errors.Error {
	var nErr error
	mq.topic, nErr = here.Here.Node().PubSub.Join(fmt.Sprintf("%s.%s", rootProtocol, mq.Topic))
	if nErr != nil {
		gLogger.Error("join topic failed", zap.Error(nErr))
		return errors.Sys("join topic error", nErr)
	}
	mq.subscription, nErr = mq.topic.Subscribe(pubsub.WithBufferSize(defaultBufferSize))
	if nErr != nil {
		gLogger.Error("subscribe topic failed", zap.Error(nErr))
		return errors.Sys("subscribe topic error", nErr)
	}
	return nil
}
