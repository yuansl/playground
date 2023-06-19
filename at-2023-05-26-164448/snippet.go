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
	"fmt"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func fatal(v ...any) {
	fmt.Fprintln(os.Stderr, "fatal error: ", v)
	os.Exit(1)
}

type Message struct {
	Header struct{}
	Body   []byte
}

type MessageQueue interface {
	Sendmsg(ctx context.Context, msg *Message) error
	Recvmsg(ctx context.Context) (*Message, error)
}

type mqueue struct {
	conn     *amqp.Connection
	ch       *amqp.Channel
	queue    string
	consumer string
}

func (mq *mqueue) Sendmsg(ctx context.Context, msg *Message) error {
	return mq.ch.PublishWithContext(ctx, "", mq.queue, false, false, amqp.Publishing{})
}

func (mq *mqueue) Recvmsg(ctx context.Context) (*Message, error) {
	c, err := mq.ch.Consume(mq.queue, mq.consumer, false, false, false, true, amqp.Table{})
	if err != nil {
		return nil, fmt.Errorf("mq.ch.Consume: %v", err)
	}
	msg := <-c
	return &Message{Body: msg.Body}, nil
}

func (mq *mqueue) Close() error {
	return mq.conn.Close()
}

func NewMessageQueue() MessageQueue {
	hostname, err := os.Hostname()
	if err != nil {
		fatal("os.Hostname:", err)
	}
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		fatal("amqp.Dial:", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		fatal("conn.Channel():", err)
	}

	return &mqueue{conn: conn, ch: ch, consumer: hostname}
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		fatal("amqp.Dial:", err)
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		fatal("conn.Channel():", err)
	}
	q, err := ch.QueueDeclare("benchmark", false, false, false, false, amqp.Table{})
	if err != nil {
		fatal("ch.QueueDeclare:", err)
	}
	start := time.Now()

	const SIZE = 10000_000_000
	for i := 0; i < SIZE; i++ {
		err = ch.PublishWithContext(context.TODO(), "", q.Name, false, false, amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte("ping"),
		})
		if err != nil {
			fatal("ch.PublishWithContext:", err)
		}
	}

	fmt.Printf("send %d messages in %.f seconds\n", SIZE, time.Since(start).Seconds())
}
