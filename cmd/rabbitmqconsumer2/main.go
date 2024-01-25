package main

import (
	"context"
	"flag"
	"time"

	"github.com/qbox/net-deftones/logger"

	"github.com/yuansl/playground/messagequeue/xrabbitmq"
	"github.com/yuansl/playground/util"
)

var fatal = util.Fatal

var _options struct {
	rabbituser     string
	rabbitpassword string
	exchange       string
	rabbitqueue    string
	rabbitaddr     string
	datetime       time.Time
	domain         string
}

func parseCmdOptions() {
	flag.StringVar(&_options.rabbituser, "user", "", "specify rabbitmq user")
	flag.StringVar(&_options.rabbitpassword, "password", "", "specify rabbitmq password")
	flag.StringVar(&_options.rabbitaddr, "addr", "amqp://defy:defy123@localhost:5672/", "specify rabbitmq address")
	flag.StringVar(&_options.rabbitqueue, "queue", "test", "speicfy rabbit queue")
	flag.StringVar(&_options.exchange, "exch", "yuansl", "specify rabbitmq exchange")
	flag.TextVar(&_options.datetime, "datetime", time.Time{}, "specify datetime of cdn log for uploading")
	flag.StringVar(&_options.domain, "domain", "www.example.com", "speicfy cdn domain for uploading log")
	flag.Parse()
}

func main() {
	var rabbitmqopts []xrabbitmq.RabbitMQOption

	parseCmdOptions()

	if _options.exchange != "" {
		rabbitmqopts = append(rabbitmqopts, xrabbitmq.WithExchange(_options.exchange, xrabbitmq.EXCHANGE_TYPE_DIRECT))
	}
	if _options.rabbitaddr != "" {
		rabbitmqopts = append(rabbitmqopts, xrabbitmq.WithAddress(_options.rabbitaddr))
	}
	if _options.rabbituser != "" {
		rabbitmqopts = append(rabbitmqopts, xrabbitmq.WithCredential(_options.rabbituser, _options.rabbitpassword))
	}
	if _options.rabbitqueue != "" {
		rabbitmqopts = append(rabbitmqopts, xrabbitmq.WithQueueName(_options.rabbitqueue))
	}
	mq, err := xrabbitmq.NewRabbitMessageQueue(rabbitmqopts...)
	if err != nil {
		util.Fatal(err)
	}
	ctx := logger.NewContext(context.Background(), logger.New())
	log := logger.FromContext(ctx)

	for {
		msg, err := mq.Recvmsg(ctx)
		if err != nil {
			util.Fatal(err)
		}
		log.Infof("\n\nReceived message: %s, messageid: %s, timestamp: %s\n", msg.Body, msg.Header.UUID, msg.Header.Timestamp)
	}

}
