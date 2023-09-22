// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-09-16 14:54:11

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"context"
	"net/http"
	"net/http/httputil"
	"time"

	"golang.org/x/net/trace"
)

func foo(ctx context.Context) {
	tracer := trace.New("fscdn.v2", "foo call")

	tracer.LazyPrintf("Calling foo ...")

	time.Sleep(1 * time.Microsecond)
}

func main() {
	trace.DebugUseAfterFinish = true

	http.HandleFunc("/v1/cdn/domains", func(w http.ResponseWriter, req *http.Request) {
		tracer := trace.New("/v1/cdn/domains", "/v1/cdn/domains")
		defer tracer.Finish()

		data, _ := httputil.DumpRequest(req, true)

		tracer.LazyPrintf("Request(raw): '%s'\n", data)

		ctx := trace.NewContext(context.TODO(), tracer)

		foo(ctx)
	})
	http.ListenAndServe(":8080", nil)
}
