package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type onSignalTerm struct {
	signal os.Signal
}

func (e onSignalTerm) Error() string {
	return e.signal.String()
}

func ContextWithSignalHandler() context.Context {
	ctx, cancel := context.WithCancelCause(context.TODO())
	go func() {
		signalq := make(chan os.Signal, 1)

		signal.Notify(signalq, os.Interrupt, os.Kill, syscall.SIGTERM)

		select {
		case sig := <-signalq:
			log.Printf("Received signal '%s', shutdown ...\n", sig)
			cancel(onSignalTerm{sig})
		}
	}()
	return ctx
}
