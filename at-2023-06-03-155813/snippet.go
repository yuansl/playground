// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-06-03 15:58:13

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

func foo[S ~[]E, E any](a S) E {
	return a[0]
}

func main() {
	b := [8]int{1, 2}
	a := &b

	fmt.Println("Results:", foo(a[:]))
}
