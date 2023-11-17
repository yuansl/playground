// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-11-17 13:24:21

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"fmt"
)

type mySlice []string

func Clone[S ~[]E, E any](s S) S {
	return append(s[:0:0], s...)
}

func main() {
	fmt.Println("Results:", Clone(mySlice{"a", "b", "c"}))
	var v1 = make(chan struct{})
	var v2 = make(chan struct{}, 1)
	v1 = v2
	_ = v1
}
