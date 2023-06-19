// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-06-18 19:22:53

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"fmt"
	"runtime"
	"strings"
	"sync/atomic"
)

func main() {
	var msgq = make(chan string, 3)
	var counter atomic.Int32

	counter.Store(3)

	go func() {
		for counter.Load() != 3 {
			runtime.Gosched()
		}
		msgq <- "main.func1"

		counter.Add(-1)
	}()
	go func() {
		for counter.Load() != 2 {
			runtime.Gosched()
		}
		msgq <- "main.func2"

		counter.Add(-1)
	}()
	go func() {
		for counter.Load() != 1 {
			runtime.Gosched()
		}
		msgq <- "main.func3"

		counter.Add(-1)

		if counter.Load() == 0 {
			close(msgq)
		}
	}()

	msgs := []string{}
	for msg := range msgq {
		msgs = append(msgs, msg)
	}
	fmt.Println(strings.Join(msgs, ","))
}
