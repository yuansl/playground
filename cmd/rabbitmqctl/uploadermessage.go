package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/qbox/net-deftones/logger"
	netutil "github.com/qbox/net-deftones/util"
	"github.com/yuansl/playground/messagequeue"
)

type MessageBody struct {
	Domain      string   `json:"domain"`
	Granularity string   `json:"granularity"`
	Timestamp   int64    `json:"timestamp"`
	Increment   bool     `json:"incrementTaskUpload"`
	OnlineCdns  []string `json:"onlineCdns"`
}

type UploaderMessage struct {
	Id   string      `json:"messageId"`
	Body MessageBody `json:"messageBody"`
}

type (
	MessageQueue = messagequeue.MessageQueue
	Message      = messagequeue.Message
	Header       = messagequeue.Header
)

type UploaderService struct {
	MessageQueue
}

func (up *UploaderService) TryUploadAsync(ctx context.Context, domain string, datetime time.Time) error {
	msg := UploaderMessage{
		Id: logger.IdFromContext(ctx),
		Body: MessageBody{
			Domain:      domain,
			Granularity: "1hour",
			Increment:   false,
			Timestamp:   datetime.Unix(),
			OnlineCdns:  []string{"all"},
		},
	}
	data, _ := json.Marshal(&msg)

	return up.Sendmsg(ctx, &Message{Header: Header{Timestamp: time.Now(), UUID: msg.Id}, Body: data})
}

type UploaderOption netutil.Option
type uploaderOptions struct {
	mqProducer MessageQueue
}

func WithMessageQueue(mq MessageQueue) UploaderOption {
	return netutil.OptionFunc(func(opt any) {
		opt.(*uploaderOptions).mqProducer = mq
	})
}

func NewUploaderMessager(opts ...UploaderOption) *UploaderService {
	var options uploaderOptions

	for _, opt := range opts {
		opt.Apply(&options)
	}
	return &UploaderService{MessageQueue: options.mqProducer}
}
