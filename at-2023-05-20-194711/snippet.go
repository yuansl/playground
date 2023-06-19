// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-05-20 19:47:11

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
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type connect struct {
	network string
	addr    string
}

func fatal(v ...any) {
	fmt.Fprintln(os.Stderr, "fatal error: ", v)
	os.Exit(1)
}

func main() {
	transp := http.DefaultTransport.(*http.Transport)
	transp.MaxConnsPerHost = 1
	transp.MaxIdleConnsPerHost = 1
	transp.MaxIdleConns = 1
	var perConnect sync.Map

	go func() {
		for range time.Tick(1 * time.Second) {
			perConnect.Range(func(key, value any) bool {
				fmt.Printf("(%s:%v)", key, value.(*atomic.Int64).Load())
				return true
			})
		}
	}()
	transp.DialContext = func(_ context.Context, network, addr string) (net.Conn, error) {
		key := connect{network, addr}
		v, exists := perConnect.Load(key)
		if !exists {
			v = new(atomic.Int64)
		}
		v.(*atomic.Int64).Add(+1)
		perConnect.Store(key, v)

		return net.Dial(network, addr)
	}
	var wg sync.WaitGroup
	for i := 0; i < 1000_000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := http.Get("https://www.example.com/")
			if err != nil {
				fatal("http.Get:", err)
			}
			defer resp.Body.Close()
			io.Copy(io.Discard, resp.Body)
		}()
	}
	wg.Wait()
}
