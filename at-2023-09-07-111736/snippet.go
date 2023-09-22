// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-09-07 11:17:36

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
	"reflect"
)

type I1[T any] interface {
	m1(T)
}
type I2[T any] interface {
	I1[T]
	m2(T)
}

var V1 I1[int]
var V2 I2[int]

func g[T any](I1[T]) {}

var v3 io.ReadCloser
var v4 io.Reader

func _() {
	g(V1)
	g(V2)
	V1 = V2

	v4 = v3
}

type T[V any] any

func f[P any](T[P]) {}

func _() {
	var x T[string]
	f[int](x)
}

type Numeric interface {
	~int | ~uint | ~int32 | ~uint32 | ~int64 | ~uint64 | ~float32 | ~float64
}

func add[T Numeric](v1, v2 T) T {
	return v1 + v2
}

type Writer[T any] interface {
	Write(T) error
}

type Writer2[T any] interface {
	Writer[T]
}

type Some struct{}

// Write implements Writer.
func (*Some) Write(byte) error {
	panic("unimplemented")
}

var _ Writer[byte] = (*Some)(nil)

func foo[T any, V Writer2[T]](_ Writer[T], _ V) {

}

func _() {
	var some Some

	foo(&some, &some)
}

func main() {
	v := 3 + 3.14 + 3
	fmt.Print("kind:", reflect.ValueOf(v).Kind())
}
