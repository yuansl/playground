package xrabbitmq

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"

	"github.com/qbox/net-deftones/logger"
	netutil "github.com/qbox/net-deftones/util"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/yuansl/playground/messagequeue"
)

type consumer struct {
	*amqp.Channel
}

type producer struct {
	*amqp.Channel
}

// rabbitmqueue implements MesssageQueue interface
type rabbitmqueue struct {
	*consumer
	*producer
	queue   string
	options *rabbitmqOptions
}

type Message = messagequeue.Message

func (mq *rabbitmqueue) Sendmsg(ctx context.Context, msg *Message) error {
	logger.FromContext(ctx).Infof("Sending message %s to queue %q ...\n", msg, mq.queue)

	return mq.producer.PublishWithContext(ctx, "", mq.queue, false, false, amqp.Publishing{
		Body:        msg.Body,
		ContentType: "application/json",
		Timestamp:   msg.Header.Timestamp,
		MessageId:   msg.Header.UUID,
	})
}

func (mq *rabbitmqueue) Recvmsg(ctx context.Context) (*Message, error) {
	c, err := mq.consumer.Consume(mq.queue, "", false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("mq.ch.Consume: %v", err)
	}
	msg := <-c
	return &Message{Body: msg.Body, Header: messagequeue.Header{UUID: msg.MessageId, Timestamp: msg.Timestamp}}, nil
}

func (mq *rabbitmqueue) Close() error {
	return errors.Join(mq.producer.Close(), mq.consumer.Close())
}

type RabbitMQOption netutil.Option

type rabbitmqOptions struct {
	addr     string
	exchange string
	exchType string
	durable  bool
	queue    string
	consumer string
	username string
	password string
}

func WithExchange(name, kind string) RabbitMQOption {
	return netutil.OptionFunc(func(opt any) {
		opt.(*rabbitmqOptions).exchange = name
		opt.(*rabbitmqOptions).exchType = kind
	})
}

func WithQueueName(queue string) RabbitMQOption {
	return netutil.OptionFunc(func(opt any) {
		opt.(*rabbitmqOptions).queue = queue
	})
}

func WithAddress(addr string) RabbitMQOption {
	return netutil.OptionFunc(func(opt any) {
		opt.(*rabbitmqOptions).addr = addr
	})
}

func WithCredential(user, password string) RabbitMQOption {
	return netutil.OptionFunc(func(opt any) {
		opt.(*rabbitmqOptions).username = user
		opt.(*rabbitmqOptions).password = password
	})
}

func WithDurable(durable bool) RabbitMQOption {
	return netutil.OptionFunc(func(opt any) {
		opt.(*rabbitmqOptions).durable = durable
	})
}

const (
	EXCHANGE_TYPE_DIRECT = "direct"
	EXCHANGE_TYPE_FANOUT = "fanout"
)

func initRabbitmqOptions(opts ...RabbitMQOption) *rabbitmqOptions {
	options := rabbitmqOptions{
		durable:  true,
		exchType: EXCHANGE_TYPE_DIRECT,
		addr:     "amqp://localhost:5672",
	}

	for _, opt := range opts {
		opt.Apply(&options)
	}
	if options.consumer == "" {
		hostname, err := os.Hostname()
		if err != nil {
			netutil.BUG(1, "os.Hostname: %v", err)
		}
		options.consumer = hostname
	}
	return &options
}

func connect(options *rabbitmqOptions) (*amqp.Connection, error) {
	return amqp.Dial(options.addr)
}

func connect2(options *rabbitmqOptions) (*amqp.Connection, error) {
	u, err := url.Parse(options.addr)
	if err != nil {
		return nil, fmt.Errorf("url.Parse(addr=%q): %v", options.addr, err)
	}
	netconn, err := net.Dial("tcp", u.Host)
	if err != nil {
		return nil, fmt.Errorf("net.Dial: %w", err)
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
	return amqp.Open(netconn, conf)
}

func initChannel(options *rabbitmqOptions) (*amqp.Channel, error) {
	conn, err := connect(options)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq.connect: %w", err)
	}
	return conn.Channel()
}

func NewRabbitmqProducer(opts ...RabbitMQOption) (*producer, error) {
	options := initRabbitmqOptions(opts...)
	ch, err := initChannel(options)

	err = ch.ExchangeDeclare(options.exchange, options.exchType, options.durable, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("channel.ExchangeDeclare: %w", err)
	}

	return &producer{Channel: ch}, nil
}

func NewRabbitmqConsumer(opts ...RabbitMQOption) (*consumer, error) {
	options := initRabbitmqOptions(opts...)
	ch, err := initChannel(options)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq.Channel: %w", err)
	}
	queue, err := ch.QueueDeclare(options.queue, options.durable, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("channel.ExchangeDeclare: %w", err)
	}
	err = ch.QueueBind(queue.Name, "", options.exchange, false, nil)
	if err != nil {
		return nil, fmt.Errorf("channel.QueueBind: %w", err)
	}
	return &consumer{Channel: ch}, nil
}

func NewRabbitMessageQueue(opts ...RabbitMQOption) (messagequeue.MessageQueue, error) {
	options := initRabbitmqOptions(opts...)
	producer, err := NewRabbitmqProducer(opts...)
	if err != nil {
		return nil, err
	}
	consumer, err := NewRabbitmqConsumer(opts...)
	if err != nil {
		return nil, err
	}
	return &rabbitmqueue{
		options:  options,
		queue:    options.queue,
		consumer: consumer,
		producer: producer,
	}, nil
}
