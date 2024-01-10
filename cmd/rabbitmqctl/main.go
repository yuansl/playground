// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-05-26 16:44:48

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"context"
	"encoding/json"
	"flag"
	"time"

	"github.com/qbox/net-deftones/logger"
	"github.com/qbox/net-deftones/util"
)

var fatal = util.Fatal

type MessageBody struct {
	Domain      string   `json:"domain"`
	Granularity string   `json:"granularity"`
	Timestamp   int64    `json:"timestamp"`
	Increment   bool     `json:"incrementTaskUpload"`
	OnlineCdns  []string `json:"onlineCdns"`
}

type UploadMessage struct {
	Id   string      `json:"messageId"`
	Body MessageBody `json:"messageBody"`
}

func Run(ctx context.Context, domain string, datetime time.Time, mq MessageQueue) {
	msg := UploadMessage{
		Id: logger.IdFromContext(ctx),
		Body: MessageBody{
			Domain:      domain,
			Granularity: "1hour",
			Increment:   false,
			Timestamp:   datetime.Unix(),
			OnlineCdns:  []string{"all"},
		},
	}

	data, _ := json.Marshal(msg)

	err := mq.Sendmsg(ctx, &Message{Header: struct {
		Timestamp time.Time
		UUID      string
	}{Timestamp: time.Now(), UUID: msg.Id}, Body: data})
	if err != nil {
		fatal("mq.Sendmsg error:", err)
	}
}

var _options struct {
	rabbituser     string
	rabbitpassword string
	rabbitqueue    string
	rabbitaddr     string
	datetime       time.Time
	domain         string
}

func parseCmdOptions() {
	flag.StringVar(&_options.rabbituser, "user", "defy", "specify rabbitmq user")
	flag.StringVar(&_options.rabbitpassword, "password", "defy#123", "specify rabbitmq password")
	flag.StringVar(&_options.rabbitaddr, "addr", "amqp://jjh712:5672/", "specify rabbitmq address")
	flag.StringVar(&_options.rabbitqueue, "queue", "unified-upload-test", "speicfy rabbit queue")
	flag.TextVar(&_options.datetime, "datetime", time.Time{}, "specify datetime of cdn log for uploading")
	flag.StringVar(&_options.domain, "domain", "qn-pcdngw.cdn.huya.com", "speicfy cdn domain for uploading log")
	flag.Parse()
}

func main() {
	parseCmdOptions()
	mq := NewRabbitMessageQueue(WithAddress(_options.rabbitaddr), WithCredential(_options.rabbituser, _options.rabbitpassword), WithQueue(_options.rabbitqueue))
	ctx := logger.NewContext(context.Background(), logger.New())

	Run(ctx, _options.domain, _options.datetime, mq)
}
