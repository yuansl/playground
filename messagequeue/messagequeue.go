package messagequeue

import (
	"context"
	"time"
)

type Header struct {
	Timestamp time.Time
	UUID      string
	Topic     string
}

type Message struct {
	Header Header
	Body   []byte
}

type MessageQueue interface {
	Sendmsg(ctx context.Context, msg *Message) error
	Recvmsg(ctx context.Context) (*Message, error)
}
