package util

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

type onSignalErr struct{ sig os.Signal }

func (e onSignalErr) Error() string {
	return fmt.Sprintf("signal handler: the signal %q received, context channel will be closed ...", e.sig.String())
}

func InitSignalHandler(ctx context.Context, onSignal ...func()) context.Context {
	ctx, cancel := context.WithCancelCause(ctx)
	go func() {
		signals := make(chan os.Signal, 1)

		signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGUSR1)
		select {
		case <-ctx.Done():
			cancel(context.Canceled)
		case sig := <-signals:
			cancel(onSignalErr{sig})
			for _, f := range onSignal {
				f()
			}
		}
	}()
	return ctx
}
