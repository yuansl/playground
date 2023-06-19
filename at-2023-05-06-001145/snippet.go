// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-05-06 00:11:45

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
	"net/http"
)

func main() {
	http.ListenAndServe(":8080", nil)
	server := http.Server{}
	server.Shutdown(context.TODO())
	fmt.Println("Results:")
}
