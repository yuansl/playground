// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-05-30 08:10:14

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"fmt"
	"reflect"
)

func initialize(v any) {
	refv := reflect.ValueOf(v)

	if refv.Kind() == reflect.Pointer {
		refv = refv.Elem()
	}
	switch refv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Slice:
		if refv.IsNil() {
			switch refv.Kind() {
			case reflect.Map:
				refv.Set(reflect.MakeMap(refv.Type()))

				for _, kv := range []struct {
					key   string
					value any
				}{
					{"name", "limint"},
					{"age", 38},
					{"address", "china,shanghai"},
				} {
					refv.SetMapIndex(reflect.ValueOf(kv.key), reflect.ValueOf(kv.value))
				}
			case reflect.Func:
				// refv.Set(reflect.MakeFunc(typ reflect.Type, fn func(args []reflect.Value) (results []reflect.Value)))
			case reflect.Chan:
				refv.Set(reflect.MakeChan(refv.Type(), 2))
				refv.Send(reflect.ValueOf(10))
				refv.Send(reflect.ValueOf(11))
				refv.Close()
			case reflect.Slice:
				slicex := reflect.MakeSlice(refv.Type(), 0, 2)
				refv.Set(reflect.Append(slicex, reflect.ValueOf(int64(2)), reflect.ValueOf(int64(3))))
			default:
			}
		}
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		refv.Set(reflect.ValueOf(int32(38)))
	default:
	}
}

func main() {
	var x map[string]any
	var y int32
	var z []int64
	var a chan int

	initialize(&x)
	initialize(&y)
	initialize(&z)
	initialize(&a)

	fmt.Println("x=", x, "y=", y, "z=", z)
	for v := range a {
		fmt.Println("Read chan:", v)
	}
}
