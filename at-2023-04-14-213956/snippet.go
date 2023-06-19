// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-04-14 21:39:56

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"fmt"
	"unsafe"
)

func main() {
	var a, b struct{}

	fmt.Printf("Results: &a=%p, &b=%p, a==b: %t\n", &a, &b, a == b)

	var c = make([]struct{}, 20)
	var d = make([]struct{}, 12)

	fmt.Printf("cap(c)=%d, sizeof(d)=%d, &c[0]:%p, &d[0]=%p, &c[0]=&d[0]: %t\n", cap(c), unsafe.Sizeof(d[0]), &c[0], &d[0], &c[0] == &d[0])
}
