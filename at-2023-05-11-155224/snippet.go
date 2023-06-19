// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-05-11 15:52:24

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"sync/atomic"
)

func fatal(v ...any) {
	fmt.Fprintln(os.Stderr, "fatal error: ", v)
	os.Exit(1)
}

func main() {
	var x atomic.Pointer[int]

	fmt.Println("load from x:", *x.Load())

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		data, err := httputil.DumpRequest(req, true)
		if err != nil {
			fatal("httputil.DumpRequest error:", err)
		}
		fmt.Printf("Received new request(raw): '%s'\n", data)
		fmt.Fprintf(w, "Hello, this is golang-driven http server\n")
	})

	// println(http.ListenAndServe(":8080", nil))
}
