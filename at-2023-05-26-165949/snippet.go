// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-05-26 16:59:49

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"fmt"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

func fatal(v ...any) {
	fmt.Fprintln(os.Stderr, "fatal error: ", v)
	os.Exit(1)
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672")
	if err != nil {
		fatal("amqp.Dial:", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		fatal("conn.Channel:", err)
	}
	q, err := ch.QueueDeclare("benchmark", false, false, false, false, amqp.Table{})
	if err != nil {
		fatal("ch.QueueDeclare:", err)
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, amqp.Table{})
	if err != nil {
		fatal("ch.Consume():", err)
	}
	for msg := range msgs {
		fmt.Printf("msg: %q\n", msg.Body)
	}
}
