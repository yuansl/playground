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

	"github.com/qbox/net-deftones/logger"
	"github.com/qbox/net-deftones/util"

	"github.com/yuansl/playground/messagequeue"
	"github.com/yuansl/playground/messagequeue/rabbitmq"
)

var fatal = util.Fatal

var options struct {
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

func parseOptions() {
	flag.StringVar(&options.rabbituser, "user", "", "specify rabbitmq user")
	flag.StringVar(&options.rabbitpassword, "password", "", "specify rabbitmq password")
	flag.StringVar(&options.rabbitaddr, "addr", "amqp://yuansl:yuansl@localhost:5672/", "specify rabbitmq address")
	flag.StringVar(&options.rabbitexch, "exch", "yuansl", "specify rabbitmq's exchange")
	flag.StringVar(&options.rabbitqueue, "queue", "test", "speicfy rabbit queue")
	flag.StringVar(&options.topic, "topic", "some", "specify topic(routing key in rabbitmq)")
	flag.TextVar(&options.datetime, "datetime", time.Time{}, "specify datetime of cdn log for uploading")
	flag.StringVar(&options.domain, "domain", "www.example.com", "speicfy cdn domain for uploading log")
	flag.StringVar(&options.role, "role", "producer", "whether this is a producer or consumer")
	flag.Parse()
}

func main() {
	parseOptions()

	var rabbitmqopts []rabbitmq.RabbitMQOption

	if options.rabbitexch != "" {
		rabbitmqopts = append(rabbitmqopts, rabbitmq.WithExchange(options.rabbitexch, rabbitmq.EXCHANGE_TYPE_DIRECT))
	}
	if options.rabbitaddr != "" {
		rabbitmqopts = append(rabbitmqopts, rabbitmq.WithAddress(options.rabbitaddr))
	}
	if options.rabbituser != "" {
		rabbitmqopts = append(rabbitmqopts, rabbitmq.WithCredential(options.rabbituser, options.rabbitpassword))
	}
	if options.rabbitqueue != "" {
		rabbitmqopts = append(rabbitmqopts, rabbitmq.WithQueueName(options.rabbitqueue))
	}
	if options.topic != "" {
		rabbitmqopts = append(rabbitmqopts, rabbitmq.WithTopic(options.topic))
	}
	if options.role != "" {
		rabbitmqopts = append(rabbitmqopts, rabbitmq.WithRole(options.role))
	}
	mq, err := rabbitmq.NewProducer(rabbitmqopts...)
	if err != nil {
		util.Fatal(err)
	}
	ctx := logger.NewContext(context.Background(), logger.New())

	err = mq.Sendmsg(ctx, &messagequeue.Message{
		Header: messagequeue.Header{
			Topic: options.topic, Timestamp: time.Now(), UUID: logger.IdFromContext(ctx),
		},
		Body: []byte("hello, this message")})
	if err != nil {
		util.Fatal(err)
	}
	fmt.Printf("done\n")
}
