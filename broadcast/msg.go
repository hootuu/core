package broadcast

import (
	"encoding/json"
	"github.com/hootuu/utils/errors"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"go.uber.org/zap"
	"strings"
)

func Encode(data interface{}) ([]byte, *errors.Error) {
	if data == nil {
		return nil, errors.Sys("payload.data is nil")
	}
	bytes, nErr := json.Marshal(data)
	if nErr != nil {
		gLogger.Error("json.Marshal failed", zap.Error(nErr))
		return nil, errors.Sys("payload encode failed")
	}
	return bytes, nil
}

func Decode[T any](data []byte) (*T, *errors.Error) {
	if len(data) == 0 {
		return nil, errors.Sys("data is empty")
	}
	var m T
	nErr := json.Unmarshal(data, &m)
	if nErr != nil {
		gLogger.Error("json.Unmarshal failed", zap.Error(nErr))
		return nil, errors.Sys("payload decode failed")
	}
	return &m, nil
}

type Message struct {
	Topic     string   `bson:"topic" json:"topic"`
	VN        string   `bson:"vn" json:"vn"`
	Scope     string   `bson:"scope" json:"scope"`
	ID        string   `bson:"id" json:"id"`
	ReplyID   string   `bson:"reply_id,omitempty" json:"reply_id"`
	From      string   `bson:"from" json:"from"`
	Data      []byte   `bson:"data" json:"data"`
	Timestamp int64    `bson:"timestamp" json:"timestamp"`
	Tag       []string `bson:"tag,omitempty" json:"tag,omitempty"`
}

type Data struct {
	VN        string   `bson:"vn" json:"vn"`
	Scope     string   `bson:"scope" json:"scope"`
	ReplyID   string   `bson:"reply_id,omitempty" json:"reply_id"`
	Data      string   `bson:"data" json:"data"`
	Timestamp int64    `bson:"timestamp" json:"timestamp"`
	Tag       []string `bson:"tag,omitempty" json:"tag,omitempty"`
}

func (data *Data) WithData(d interface{}) *errors.Error {
	enBytes, err := Encode(d)
	if err != nil {
		return err
	}
	data.Data = string(enBytes)
	return nil
}

func (data *Data) ToBytes() ([]byte, *errors.Error) {
	enBytes, err := Encode(data)
	if err != nil {
		return nil, err
	}
	return enBytes, nil
}

func DataOfBytes(payload []byte) (*Data, *errors.Error) {
	if len(payload) == 0 {
		return nil, errors.Sys("invalid payload, it is nil")
	}
	var data Data
	nErr := json.Unmarshal(payload, &data)
	if nErr != nil {
		gLogger.Error("json.Unmarshal failed", zap.Error(nErr))
		return nil, errors.Sys("invalid payload:"+nErr.Error(), nErr)
	}
	return &data, nil
}

func MessageOf(pbMsg *pubsub.Message) (*Message, *errors.Error) {
	innerData, err := DataOfBytes(pbMsg.GetData())
	if err != nil {
		return nil, err
	}
	msg := &Message{
		Topic:     pbMsg.GetTopic(),
		VN:        innerData.VN,
		Scope:     innerData.Scope,
		ID:        pbMsg.ID,
		ReplyID:   innerData.ReplyID,
		From:      pbMsg.GetFrom().String(),
		Data:      []byte(innerData.Data),
		Timestamp: innerData.Timestamp,
		Tag:       innerData.Tag,
	}
	return msg, nil
}

func (msg *Message) Summary() string {
	return strings.Join([]string{
		msg.From, msg.VN, msg.Scope,
	}, "|")
}
