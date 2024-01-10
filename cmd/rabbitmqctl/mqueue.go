package main

import (
	"context"
	"time"
)

type Message struct {
	Header struct {
		Timestamp time.Time
		UUID      string
	}
	Body []byte
}

type MessageQueue interface {
	Sendmsg(ctx context.Context, msg *Message) error
	Recvmsg(ctx context.Context) (*Message, error)
}
