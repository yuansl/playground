package main

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/qbox/net-deftones/logger"
	"github.com/yuansl/playground/util"
)

// rabbitmqueue implements MesssageQueue interface
type rabbitmqueue struct {
	conn    *amqp.Connection
	ch      *amqp.Channel
	queue   string
	options *options
}

func (mq *rabbitmqueue) Sendmsg(ctx context.Context, msg *Message) error {
	logger.FromContext(ctx).Infof("Sending message %s ...\n", msg)

	return mq.ch.PublishWithContext(ctx, "", mq.queue, false, false, amqp.Publishing{
		Body:        msg.Body,
		ContentType: "application/json",
		Timestamp:   msg.Header.Timestamp,
		MessageId:   msg.Header.UUID,
	})
}

func (mq *rabbitmqueue) Recvmsg(ctx context.Context) (*Message, error) {
	c, err := mq.ch.Consume(mq.queue, mq.options.consumer, false, false, false, true, amqp.Table{})
	if err != nil {
		return nil, fmt.Errorf("mq.ch.Consume: %v", err)
	}
	msg := <-c
	return &Message{Body: msg.Body}, nil
}

func (mq *rabbitmqueue) Close() error {
	return mq.conn.Close()
}

type Option util.Option

type options struct {
	addr     string
	queue    string
	consumer string
	username string
	password string
}

func WithQueue(queue string) Option {
	return util.OptionFunc(func(opt any) {
		opt.(*options).queue = queue
	})
}

func WithAddress(addr string) Option {
	return util.OptionFunc(func(opt any) {
		opt.(*options).addr = addr
	})
}

func WithCredential(user, password string) Option {
	return util.OptionFunc(func(opt any) {
		opt.(*options).username = user
		opt.(*options).password = password
	})
}

func NewRabbitMessageQueue(opts ...Option) MessageQueue {
	var options options

	for _, opt := range opts {
		opt.Apply(&options)
	}
	if options.consumer == "" {
		hostname, err := os.Hostname()
		if err != nil {
			fatal("os.Hostname:", err)
		}
		options.consumer = hostname
	}
	if options.addr == "" {
		options.addr = "localhost:5672"
	}
	u, err := url.Parse(options.addr)
	if err != nil {
		fatal("url.Parse(addr=%q): %v\n", options.addr, err)
	}
	netconn, err := net.Dial("tcp", u.Host)
	if err != nil {
		fatal("net.Dial:", err)
	}
	conf := amqp.Config{
		Dial: func(network, addr string) (net.Conn, error) {
			fmt.Printf("Dialing address: '%s/%s' ...\n", network, addr)
			return net.Dial(network, addr)
		},
		Vhost: "/",
	}
	if options.username != "" {
		conf.SASL = []amqp.Authentication{
			&amqp.AMQPlainAuth{
				Username: options.username, Password: options.password,
			}}
	}
	conn, err := amqp.Open(netconn, conf)
	if err != nil {
		fatal("amqp.Open:", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		fatal("conn.Channel():", err)
	}
	return &rabbitmqueue{conn: conn, ch: ch, options: &options, queue: options.queue}
}
