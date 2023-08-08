// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-06-21 14:04:59

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"fmt"
	"io"
	"unsafe"
)

type Writer interface {
	io.Writer
	io.WriteCloser
}
type T int
type S struct{ *T }

func (t T) M() { fmt.Printf("t = %d\n", t) }

func main() {
	var x = []byte{0x01, 0x02, 0x03, 0x04, 06, 07, 010, 011, 012, 013, 014, 015, 016}
	var y = *(*[2]byte)(x)
	_ = y

	t := new(T)
	s := S{t}

	*t = 1

	f := t.M
	g := s.M
	*t = 2

	f()
	g()
	fmt.Printf("sizeof(int)=%d\n", unsafe.Sizeof(int(0)))
}
