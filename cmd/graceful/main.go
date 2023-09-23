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
	"strconv"
	"sync"
	"syscall"

	"github.com/yuansl/playground/utils"
)

var HOSTNAME string

func init() {
	sync.OnceFunc(func() {
		var err error
		if HOSTNAME, err = os.Hostname(); err != nil {
			fatal("os.Hostname error:", err)
		}
	})()
}

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

var (
	graceful bool

	fatal = utils.Fatal
)

func parseCmdArgs() {
	flag.BoolVar(&graceful, "graceful", false, "graceful reload the service(process)")
	flag.Parse()
}

func reload(ln net.Listener) {
	fmt.Println("Recevied reload signal, the service will be reloaded as soon as possible ...")

	if tcplistener, ok := ln.(*net.TCPListener); ok {
		listenfd, err := tcplistener.File()
		if err != nil {
			fatal("tcplistener.File() error:", err)
		}

		cmd := exec.Command(os.Args[0], "-graceful", strconv.FormatBool(true))

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.ExtraFiles = []*os.File{listenfd}

		if err := cmd.Start(); err != nil {
			fatal("cmd.Start error:", err)
		}
		fmt.Printf("child process started as pid @%d\n", cmd.Process.Pid)
	}
}

func run(ctx context.Context, server *http.Server, ln net.Listener) error {
	errchan := make(chan error, 1)

	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
			data, _ := httputil.DumpRequest(req, true)
			log.Printf("Request(raw): '%s'\n", data)
			w.Write([]byte(`Hello, this is process ` + strconv.Itoa(os.Getpid()) + " at host @" + HOSTNAME))
		})

		errchan <- server.Serve(ln)
	}()
	select {
	case <-ctx.Done():
		switch err := ctx.Err(); {
		case errors.Is(err, context.Canceled):
			var cause *onSignal

			if errors.As(context.Cause(ctx), &cause) {
				switch cause.sig {
				case syscall.SIGUSR1:
					reload(ln)
				default:
					// default action on signal
				}
			}
			server.Shutdown(context.WithoutCancel(ctx))
		default:
			return err
		}
	case err := <-errchan:
		log.Printf("Recevied error: %v, shutdown ...\n", err)
		server.Shutdown(ctx)
	}
	return nil
}

func NumFilesLimit() int {
	var rlim syscall.Rlimit

	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlim); err != nil {
		fatal("syscall.Getrlimit:", err)
	}
	return int(rlim.Max)
}

// findFirstOpenFile finds the first read/writeable listen fd
func findFirstOpenFile() *os.File {
	for proc := os.Stderr.Fd() + 1; proc < uintptr(NumFilesLimit()); proc++ {
		if f := os.NewFile(uintptr(proc), ""); f != nil {
			return f
		}
	}
	return nil
}

func main() {
	parseCmdArgs()

	var (
		ln  net.Listener
		err error
	)
	if graceful {
		ln, err = net.FileListener(findFirstOpenFile())
	} else {
		ln, err = net.Listen("tcp", ":8080")
	}
	if err != nil {
		fatal("Listen addr error:", err)
	}

	ctx := InitSignalHandler()

	if err := run(ctx, &http.Server{}, ln); err != nil {
		fatal("run error: ", err)
	}
}
