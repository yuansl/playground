// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-07-28 11:01:20

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"cmp"
	"io"
	"os"
	"reflect"
	"sync"

	"golang.org/x/exp/constraints"
)

type Some[T any] struct {
	Value T
}

func (s Some[T]) get() T {
	return s.Value
}

type MyWriter io.Writer

func Write(w MyWriter) {
	w.Write(nil)
}

type PrintableMutext struct {
	sync.Mutex
}

func pretty(x *PrintableMutext) {
	x.Lock()
}

type Person struct {
	Age     int
	Address string
}

func IndexFunc[S ~[]E, E any](s S, predicate func(E) bool) int {
	for i, v := range s {
		if predicate(v) {
			return i
		}
	}
	return -1
}

func Predicate[E constraints.Integer](v E) bool {
	return v%2 == 0
}

func ExampleGenericFunc() {
	IndexFunc([]int{1, 2, 3, 4, 5}, Predicate)
}

type UniverseWriter[T any] interface {
	Write(T) (int, error)
}

func WriteAny[T any](w UniverseWriter[T], v T) {
	w.Write(v)
}

func ExampleInferFromInterfaceMethod() {
	reflect.TypeOf(nil)
	WriteAny((*os.File)(nil), []byte("some")) // inferred UniverseWriter[[]byte] as Writer.Write([]byte)(int,error)
}

func sum[E cmp.Ordered](v ...E) E {
	var s E
	for _, vv := range v {
		s += vv
	}
	return s
}

func ExampleInferUntypedConstant() {
	sum(1, 2.0, 3) // outputs 6.0
}

func ExampleComponents[T any]([]T) {}

type Setter2[P1 any] interface {
	Set(string)
	*P1
}

func FromStrings2[P2 any, P3 Setter2[P2]]([]string) []P2 {
	return nil
}

type Settable int

func (p *Settable) Set(s string) {}

func ExampleFromStrings() {
	FromStrings2[Settable]([]string{"1"})
}

func foo2[K interface {
	comparable
	any
}]() {
}

func main() {
	ExampleComponents([]int{})

	foo2[any]()
}
