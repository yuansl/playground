// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-10-14 12:05:35

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/yuansl/playground/util"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		data, _ := httputil.DumpRequest(req, true)
		log.Printf("Request: '%s'\n", data)
	})
	if err := http.ListenAndServe(":8080", nil); err != nil {
		util.Fatal(err)
	}
}
