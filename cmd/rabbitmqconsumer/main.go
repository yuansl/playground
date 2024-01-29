package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/qbox/net-deftones/logger"

	"github.com/yuansl/playground/messagequeue/rabbitmq"
	"github.com/yuansl/playground/util"
)

var _options struct {
	rabbituser     string
	rabbitpassword string
	exchange       string
	rabbitqueue    string
	rabbitaddr     string
	datetime       time.Time
	domain         string
	topic          string
	role           string
	consumer       string
}

func parseCmdOptions() {
	flag.StringVar(&_options.rabbituser, "user", "", "specify rabbitmq user")
	flag.StringVar(&_options.rabbitpassword, "password", "", "specify rabbitmq password")
	flag.StringVar(&_options.rabbitaddr, "addr", "amqp://yuansl:yuansl@localhost:5672/", "specify rabbitmq address")
	flag.StringVar(&_options.rabbitqueue, "queue", "test", "speicfy rabbit queue")
	flag.StringVar(&_options.exchange, "exch", "yuansl", "specify rabbitmq exchange")
	flag.StringVar(&_options.topic, "topic", "some", "specify topic(routing key in rabbitmq)")
	hostname, _ := os.Hostname()
	flag.StringVar(&_options.consumer, "consumer", hostname+"001", "specify consumer tag")
	flag.TextVar(&_options.datetime, "datetime", time.Time{}, "specify datetime of cdn log for uploading")
	flag.StringVar(&_options.role, "role", "producer", "whether this is a producer or consumer")
	flag.StringVar(&_options.domain, "domain", "www.example.com", "speicfy cdn domain for uploading log")
	flag.Parse()
}

func main() {
	var rabbitmqopts []rabbitmq.RabbitMQOption

	parseCmdOptions()

	if _options.exchange != "" {
		rabbitmqopts = append(rabbitmqopts, rabbitmq.WithExchange(_options.exchange, rabbitmq.EXCHANGE_TYPE_DIRECT))
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
	if _options.consumer != "" {
		rabbitmqopts = append(rabbitmqopts, rabbitmq.WithConsumer(_options.consumer))
	}
	mq, err := rabbitmq.NewConsumer(rabbitmqopts...)
	if err != nil {
		util.Fatal(err)
	}
	ctx := logger.NewContext(context.Background(), logger.New())
	log := logger.FromContext(ctx)

	var counter atomic.Int32

	start := time.Now()

	go func() {
		for range time.Tick(1 * time.Second) {
			fmt.Printf("received %d messages after %v\n", counter.Load(), time.Since(start))
		}
	}()

	for {
		msg, err := mq.Recvmsg(ctx)
		if err != nil {
			util.Fatal(err)
		}

		counter.Add(+1)

		log.Infof("\n\nReceived message from {exchange=%s, queue=%s, key=%s}: '%s', messageid: %s, timestamp: %s\n",
			_options.exchange, _options.rabbitqueue, _options.topic,
			msg.Body, msg.Header.UUID, msg.Header.Timestamp)
	}
}
