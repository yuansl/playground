// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-10-08 13:12:27

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"fmt"
	"sync/atomic"
	"unsafe"
)

func main() {
	var x = 36
	var p *int = &x
	var unsafep = unsafe.Pointer(p)
	var unsafep2 atomic.Pointer[int]
	unsafep2.Store(p)

	atomic.StorePointer(&unsafep, unsafe.Pointer(p))

	var pp *int = (*int)(atomic.LoadPointer(&unsafep))

	fmt.Println("Results:", *pp, p, *unsafep2.Load())
}
