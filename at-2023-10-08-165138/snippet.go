// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-10-08 16:51:38

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println("Results:", strings.SplitN("localhost", ":", 2))
}
