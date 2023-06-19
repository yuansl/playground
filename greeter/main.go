package main

import (
	"context"
	"fmt"
	"os"
)

type Response struct {
	message string
}

type Request struct {
	message string
	resp    *Response
}

func greet(ctx context.Context, greeting string) (*Response, error) {
	requestQ := fromContext(ctx)
	resp := Response{}
	requestQ <- Request{message: greeting, resp: &resp}
	return &resp, nil
}

func startGreeter(ctx context.Context) {
	queue := fromContext(ctx)
	for e := range queue {
		e.resp.message = e.message + ", world"
	}
}

func main() {
	ctx := withValue(context.Background(), make(chan Request))
	go startGreeter(ctx)
	resp, err := greet(ctx, "hello")
	if err != nil {
		fatal("greet error: %v\n", err)
	}
	fmt.Printf("greet response: %+v\n", resp)
}

type key int

var ctxKey key

func withValue(ctx context.Context, value chan Request) context.Context {
	return context.WithValue(ctx, ctxKey, value)
}

func fromContext(ctx context.Context) chan Request {
	if v, ok := ctx.Value(ctxKey).(chan Request); !ok {
		panic("ctx does not contain a `chan greeterRequest`")
	} else {
		return v
	}
}

func fatal(format string, v ...any) {
	fmt.Fprintf(os.Stderr, format, v...)
	os.Exit(1)
}
