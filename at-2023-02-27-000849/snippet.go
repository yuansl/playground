// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-02-27 00:08:49

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"unsafe"
)

func main() {
	r := io.LimitReader(bytes.NewReader([]byte("abcdefghijklmnopqrst")), 10)
	r0 := bufio.NewReader(r)
	n, err := r0.ReadBytes('\n')
	fmt.Println("Results: n=", *(*string)(unsafe.Pointer(&n)), "err=", err)
}
