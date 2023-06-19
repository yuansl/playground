// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-03-12 10:30:34

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
)

func main() {
	http.ListenAndServe(":8080", nil)
	fmt.Println("Results:")
}
