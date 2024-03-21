package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/qbox/net-deftones/logger"
	"github.com/qbox/net-deftones/util"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/yuansl/playground/messagequeue"
)

type Message struct {
	*messagequeue.Message
	Queue string
}

type producer struct {
	*amqp.Channel
	exchange string
	queue    string
}

func (mqp *producer) Sendmsg(ctx context.Context, msg *messagequeue.Message) error {
	return mqp.Sendmsg1(ctx, &Message{msg, mqp.queue})
}

func (mqp *producer) Sendmsg1(ctx context.Context, msg *Message) error {
	logger.FromContext(ctx).Infof("Sending message to {exchange=%s, queue=%q, routingkey=%q}: '%s' ...\n",
		mqp.exchange, msg.Queue, msg.Header.Topic, msg)

	return mqp.PublishWithContext(ctx, mqp.exchange, msg.Header.Topic, false, false, amqp.Publishing{
		Body:        msg.Body,
		ContentType: "application/json",
		Timestamp:   msg.Header.Timestamp,
		MessageId:   msg.Header.UUID,
	})
}

func NewProducer(opts ...RabbitMQOption) (*producer, error) {
	options := initRabbitmqOptions(opts...)
	ch, err := initChannel(options)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq.Channel: %w", err)
	}
	err = ch.ExchangeDeclare(options.exchange, options.exchType, options.durable, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("channel.ExchangeDeclare: %w", err)
	}

	return &producer{Channel: ch, exchange: options.exchange, queue: options.queue}, nil
}

type consumer struct {
	*amqp.Channel
	queue   string
	mq      chan *Message
	options *rabbitmqOptions
}

func (mqc *consumer) Recvmsg1(ctx context.Context) (*Message, error) {
	msg, ok := <-mqc.mq
	if !ok {
		return nil, fmt.Errorf("mq closed")
	}
	return msg, nil
}

func (mqc *consumer) Recvmsg(ctx context.Context) (*messagequeue.Message, error) {
	msg, err := mqc.Recvmsg1(ctx)
	if err != nil {
		return nil, err
	}
	return msg.Message, nil
}

func (mqc *consumer) consume(_ context.Context) error {
	c, err := mqc.Consume(mqc.queue, mqc.options.consumer, false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("mq.ch.Consume: %v", err)
	}
	for msg := range c {
		if err = msg.Ack(true); err != nil {
			return fmt.Errorf("rabbitmq.Ack: %w", err)
		}
		mqc.mq <- &Message{Message: &messagequeue.Message{Body: msg.Body, Header: messagequeue.Header{UUID: msg.MessageId, Timestamp: msg.Timestamp}}}
	}
	return nil
}

func NewConsumer(opts ...RabbitMQOption) (*consumer, error) {
	options := initRabbitmqOptions(opts...)
	ch, err := initChannel(options)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq.Channel: %w", err)
	}
	queue, err := ch.QueueDeclare(options.queue, options.durable, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("channel.ExchangeDeclare: %w", err)
	}
	err = ch.QueueBind(queue.Name, options.topic, options.exchange, false, nil)
	if err != nil {
		return nil, fmt.Errorf("channel.QueueBind: %w", err)
	}

	c := &consumer{
		Channel: ch,
		queue:   options.queue,
		mq:      make(chan *Message, runtime.NumCPU()),
		options: options,
	}

	go func() {
		err = c.consume(context.Background())
		if err != nil {
			util.Fatal(err)
		}
	}()

	return c, nil
}

// rabbitmqueue implements MesssageQueue interface
type rabbitmqueue struct {
	*consumer
	*producer
	options *rabbitmqOptions
}

func (mq *rabbitmqueue) Close() error {
	var err error
	if mq.producer != nil {
		err = mq.producer.Close()
	}
	if mq.consumer != nil {
		err = errors.Join(err, mq.consumer.Close())
	}
	return err
}

type RabbitMQOption util.Option

type rabbitmqOptions struct {
	addr     string
	topic    string
	role     string
	exchange string
	exchType string
	durable  bool
	queue    string
	consumer string
	username string
	password string
}

func WithConsumer(consumer string) RabbitMQOption {
	return util.OptionFunc(func(opt any) {
		opt.(*rabbitmqOptions).consumer = consumer
	})
}

func WithTopic(topic string) RabbitMQOption {
	return util.OptionFunc(func(opt any) {
		opt.(*rabbitmqOptions).topic = topic
	})
}

func WithRole(role string) RabbitMQOption {
	return util.OptionFunc(func(opt any) {
		opt.(*rabbitmqOptions).role = role
	})
}

func WithExchange(name, kind string) RabbitMQOption {
	return util.OptionFunc(func(opt any) {
		opt.(*rabbitmqOptions).exchange = name
		opt.(*rabbitmqOptions).exchType = kind
	})
}

func WithQueueName(queue string) RabbitMQOption {
	return util.OptionFunc(func(opt any) {
		opt.(*rabbitmqOptions).queue = queue
	})
}

func WithAddress(addr string) RabbitMQOption {
	return util.OptionFunc(func(opt any) {
		opt.(*rabbitmqOptions).addr = addr
	})
}

func WithCredential(user, password string) RabbitMQOption {
	return util.OptionFunc(func(opt any) {
		opt.(*rabbitmqOptions).username = user
		opt.(*rabbitmqOptions).password = password
	})
}

func WithDurable(durable bool) RabbitMQOption {
	return util.OptionFunc(func(opt any) {
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
			util.BUG(1, "os.Hostname: %v", err)
		}
		options.consumer = hostname
	}
	return &options
}

func initChannel(options *rabbitmqOptions) (*amqp.Channel, error) {
	conn, err := connect(options)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq.connect: %w", err)
	}
	return conn.Channel()
}

func NewMessageQueue(opts ...RabbitMQOption) (messagequeue.MessageQueue, error) {
	options := initRabbitmqOptions(opts...)
	var (
		err      error
		producer *producer
		consumer *consumer
	)
	if options.role == "producer" {
		producer, err = NewProducer(opts...)
		if err != nil {
			return nil, err
		}
	} else {
		consumer, err = NewConsumer(opts...)
		if err != nil {
			return nil, err
		}
	}
	return &rabbitmqueue{
		options:  options,
		consumer: consumer,
		producer: producer,
	}, nil
}
