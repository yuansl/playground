// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-08-28 16:10:53

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	_RABBITMQ_ADDRESS = "localhost:5672"
	_RABBITMQ_QUEUE   = "unified-upload-v2-kodofs"
)

var (
	hostname         string
	callHostnameOnce sync.Once
)

func init() {
	callHostnameOnce.Do(func() {
		hostname, _ = os.Hostname()
	})
}

func fatal(v ...any) {
	pc, file, line, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)

	fmt.Fprintln(os.Stderr, "fatal error: ", v)
	fmt.Fprintf(os.Stderr, "%s:\n\t%s:%d\n", fn.Name(), file, line)
	os.Exit(1)
}

type rabbitmqClient struct {
	*amqp.Channel
	clientId string
}

func (client *rabbitmqClient) GetQueue(ctx context.Context, queue string) (<-chan amqp.Delivery, error) {
	select {
	case <-ctx.Done():
		switch {
		case errors.Is(ctx.Err(), context.Canceled):
			return nil, context.Cause(ctx)
		default:
			return nil, ctx.Err()
		}
	default:
	}

	return client.Channel.Consume(queue, client.clientId, false, false, false, false, nil)
}

type RabbitmqOption Option

type rabbitmqOptions struct {
	credential *amqp.PlainAuth
	address    string
	vhost      string
	clientId   string
}

func WithCredential(user, password string) RabbitmqOption {
	return OptionFunc(func(op any) {
		op.(*rabbitmqOptions).credential = &amqp.PlainAuth{Username: user, Password: password}
	})
}

func WithRabbitmqAddress(addr string) RabbitmqOption {
	return OptionFunc(func(op any) {
		op.(*rabbitmqOptions).address = addr
	})
}

func WithVhost(vhost string) RabbitmqOption {
	return OptionFunc(func(op any) {
		op.(*rabbitmqOptions).vhost = vhost
	})
}

func WithClientId(id string) RabbitmqOption {
	return OptionFunc(func(op any) {
		op.(*rabbitmqOptions).clientId = id
	})
}

func NewRabbitmqClient(opts ...RabbitmqOption) *rabbitmqClient {
	var options rabbitmqOptions

	for _, op := range opts {
		op.Apply(&options)
	}
	if options.address == "" {
		fatal("rabbitmq address must not be empty")
	}
	conn, err := net.Dial("tcp", options.address)
	if err != nil {
		fatal("net.Dial:", err)
	}

	conf := amqp.Config{Vhost: options.vhost}

	if options.credential != nil {
		conf.SASL = append(conf.SASL, options.credential)
	}
	rbconn, err := amqp.Open(conn, conf)
	if err != nil {
		fatal("amqp.Open:", err)
	}
	channel, err := rbconn.Channel()
	if err != nil {
		fatal("amqp.Channel:", err)
	}
	if options.clientId == "" {
		options.clientId = hostname
	}
	return &rabbitmqClient{Channel: channel, clientId: options.clientId}
}

var (
	username         string
	password         string
	rabbitmq_address string
	rabbitmq_vhost   string
	rabbitmq_queue   string
)

func parseCmdArgs() {
	flag.StringVar(&username, "user", os.Getenv("RABBITMQ_USER"), "specify username of rabbitmq")
	flag.StringVar(&password, "password", os.Getenv("RABBITMQ_PASSWORD"), "specify password of rabbitmq")
	flag.StringVar(&rabbitmq_address, "addr", _RABBITMQ_ADDRESS, "specify rabbitmq address")
	flag.StringVar(&rabbitmq_vhost, "vhost", "/", "specify rabbitmq vhost")
	flag.StringVar(&rabbitmq_queue, "queue", _RABBITMQ_QUEUE, "speicfy rabbitmq queue")
	flag.Parse()
}

func main() {
	parseCmdArgs()

	ctx := ContextWithSignalHandler()
	client := NewRabbitmqClient(
		WithCredential(username, password),
		WithRabbitmqAddress(rabbitmq_address),
		WithVhost(rabbitmq_vhost),
	)
	mq, err := client.GetQueue(ctx, rabbitmq_queue)
	if err != nil {
		fatal("client.Consume:", err)
	}

	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ctx.Done():
			log.Printf("context closed: %v\n", ctx.Err())
			goto done
		case msg := <-mq:
			log.Printf("Received message: '%+v', body=%q\n", msg, msg.Body)
		case <-ticker.C:
			log.Println("tick, tock!")
		}
	}
done:
	log.Println("Server shutdown!")
}
