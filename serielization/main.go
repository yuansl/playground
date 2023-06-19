package main

import (
	"fmt"
	"reflect"
)

func main() {
	var x struct {
		D string
		C int
		B int64
		A byte
	}
	refv := reflect.ValueOf(x)

	for i := 0; i < refv.NumField(); i++ {
		refv.Type().Len()
		fmt.Printf("#%d field: %#v\n", i, refv.Field(i))
	}
}
