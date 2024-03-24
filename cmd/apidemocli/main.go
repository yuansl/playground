package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/qbox/net-deftones/logger"
	"github.com/qbox/net-deftones/util"

	"github.com/yuansl/playground/trace"
)

const _REQUEST_TIMEOUT_DEFAULT = 30 * time.Millisecond

var (
	_N               int
	_concurrency     int
	_numSentRequests atomic.Int64
)

func parseCmdArgs() {
	flag.IntVar(&_N, "n", 1, "number of requests")
	flag.IntVar(&_concurrency, "c", runtime.NumCPU(), "concurrency")
	flag.Parse()
}

func doListUsers(ctx context.Context, acct AccountService) ([]User, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	res, err := acct.ListUser(ctx, 999999)
	if err != nil {
		switch {
		case errors.Is(err, ErrUnknown):
			util.Fatal(err)
		default:
			return nil, fmt.Errorf("Something went wrong: '%v'\n", err)
		}
	}

	return res, nil
}

func run(ctx context.Context, acct AccountService) error {
	ctx, cancel := context.WithCancelCause(ctx)
	errorq := make(chan error, 1)

	go func() {
		defer close(errorq)

		var climit = make(chan struct{}, _concurrency)
		var wg sync.WaitGroup

		for i := 0; i < _N; i++ {
			climit <- struct{}{}
			wg.Add(1)
			go func() {
				defer func() {
					<-climit
					wg.Done()
				}()

				ctx, cancel := context.WithTimeout(ctx, _REQUEST_TIMEOUT_DEFAULT)
				defer cancel()

				_numSentRequests.Add(+1)

				ctx, span := trace.GetTracerProvider().Tracer("").Start(ctx, "func1")
				defer span.End()

				res, err := doListUsers(ctx, acct)
				if err != nil {
					errorq <- err
					return
				}
				logger.NewWith(span.SpanContext().TraceID().String()).Infof("Response=%+v\n", res)
			}()
		}
		wg.Wait()
	}()

	for {
		select {
		case <-ctx.Done():
			switch err := ctx.Err(); {
			case errors.Is(err, context.Canceled):
				return context.Cause(ctx)
			default:
				return err
			}
		case err := <-errorq:
			if err != nil {
				cancel(err)
				return err
			}
		}
	}
}

func main() {
	parseCmdArgs()

	client := NewAccountService(":10010")

	start := time.Now()
	defer func() { fmt.Printf("send %d requests, time elapsed: %v\n", _numSentRequests.Load(), time.Since(start)) }()

	ctx := util.InitSignalHandler(logger.NewContext(context.TODO(), logger.New()))

	run(ctx, client)
}
