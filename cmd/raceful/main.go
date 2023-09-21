package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/yuansl/playground/utils"
)

type onSignal struct {
	sig os.Signal
}

// Error implements error.
func (e *onSignal) Error() string {
	return fmt.Sprintf("The signal %v received, shutdown ...", e.sig)
}

var _ error = (*onSignal)(nil)

func InitSignalHandler() context.Context {
	ctx, cancel := context.WithCancelCause(context.TODO())
	go func() {
		sigchan := make(chan os.Signal, 1)

		signal.Notify(sigchan, syscall.SIGUSR1, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-ctx.Done():
			switch err := ctx.Err(); {
			case errors.Is(err, context.Canceled):
				cancel(context.Cause(ctx))
			default:
				cancel(err)
			}
		case sig := <-sigchan:
			cancel(&onSignal{sig})
		}
	}()
	return ctx
}

var graceful int

func parseCmdArgs() {
	flag.IntVar(&graceful, "graceful", -1, "if start the process gracefully")
	flag.Parse()
}

func reload(fd int) {
	ln, err := net.FileListener(os.NewFile(uintptr(fd), ""))
	if err != nil {
		utils.Fatal(err)
	}
}

func run(ctx context.Context, server *http.Server) error {
	var errchan = make(chan error, 1)

	go func() {
		errchan <- server.ListenAndServe()
	}()
	select {
	case <-ctx.Done():
		switch err := ctx.Err(); {
		case errors.Is(err, context.Canceled):
			err = context.Cause(ctx)
			var cause *onSignal
			if errors.As(err, &cause) && cause.sig == syscall.SIGUSR1 {
				// reload(server.)
			}
			server.Shutdown(ctx)
		default:
			return err
		}
	case err := <-errchan:
		log.Printf("Recevied error: %v, shutdown ...\n", err)
		server.Shutdown(ctx)
	}
	return nil
}

func main() {
	parseCmdArgs()

	if graceful {
		cmd := exec.Command(os.Args[0], os.Args[1:]...)
		cmd.ExtraFiles = []*os.File{os.NewFile(3, "")}
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Start()
	}

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		data, _ := httputil.DumpRequest(req, true)
		log.Printf("Request(raw): '%s'\n", data)
	})

	server := http.Server{
		Addr: ":8080",
	}
	ctx := InitSignalHandler()
	fmt.Println(run(ctx, &server))
}
