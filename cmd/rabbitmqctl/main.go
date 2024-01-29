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
	"flag"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/qbox/net-deftones/logger"
	"github.com/qbox/net-deftones/util"
	"github.com/yuansl/playground/messagequeue"
	"github.com/yuansl/playground/messagequeue/rabbitmq"
)

var fatal = util.Fatal

var _options struct {
	rabbituser     string
	rabbitpassword string
	rabbitexch     string
	rabbitqueue    string
	rabbitaddr     string
	datetime       time.Time
	domain         string
	topic          string
	role           string
}

func parseCmdOptions() {
	flag.StringVar(&_options.rabbituser, "user", "", "specify rabbitmq user")
	flag.StringVar(&_options.rabbitpassword, "password", "", "specify rabbitmq password")
	flag.StringVar(&_options.rabbitaddr, "addr", "amqp://yuansl:yuansl@localhost:5672/", "specify rabbitmq address")
	flag.StringVar(&_options.rabbitexch, "exch", "yuansl", "specify rabbitmq's exchange")
	flag.StringVar(&_options.rabbitqueue, "queue", "test", "speicfy rabbit queue")
	flag.StringVar(&_options.topic, "topic", "some", "specify topic(routing key in rabbitmq)")
	flag.TextVar(&_options.datetime, "datetime", time.Time{}, "specify datetime of cdn log for uploading")
	flag.StringVar(&_options.domain, "domain", "www.example.com", "speicfy cdn domain for uploading log")
	flag.StringVar(&_options.role, "role", "producer", "whether this is a producer or consumer")
	flag.Parse()
}

func main() {
	parseCmdOptions()

	var rabbitmqopts []rabbitmq.RabbitMQOption

	if _options.rabbitexch != "" {
		rabbitmqopts = append(rabbitmqopts, rabbitmq.WithExchange(_options.rabbitexch, rabbitmq.EXCHANGE_TYPE_DIRECT))
	}
	if _options.rabbitaddr != "" {
		rabbitmqopts = append(rabbitmqopts, rabbitmq.WithAddress(_options.rabbitaddr))
	}
	if _options.rabbituser != "" {
		rabbitmqopts = append(rabbitmqopts, rabbitmq.WithCredential(_options.rabbituser, _options.rabbitpassword))
	}
	if _options.rabbitqueue != "" {
		rabbitmqopts = append(rabbitmqopts, rabbitmq.WithQueueName(_options.rabbitqueue))
	}
	if _options.topic != "" {
		rabbitmqopts = append(rabbitmqopts, rabbitmq.WithTopic(_options.topic))
	}
	if _options.role != "" {
		rabbitmqopts = append(rabbitmqopts, rabbitmq.WithRole(_options.role))
	}
	mq, err := rabbitmq.NewProducer(rabbitmqopts...)
	if err != nil {
		util.Fatal(err)
	}
	ctx := logger.NewContext(context.Background(), logger.NewWith(uuid.NewString()))

	err = mq.Sendmsg(ctx, &messagequeue.Message{
		Header: messagequeue.Header{
			Queue: _options.rabbitqueue,
			Topic: _options.topic, Timestamp: time.Now(), UUID: logger.IdFromContext(ctx),
		},
		Body: []byte("hello, this message")})
	if err != nil {
		util.Fatal(err)
	}
	fmt.Printf("done\n")
}
