package main

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
)

type nopReader struct{}

func (nopReader) Read(p []byte) (int, error) {
	return 0, io.EOF
}

func foo() {
	var r = &bytes.Buffer{}

	typeofreader := reflect.TypeOf((*io.Reader)(nil)).Elem()

	inspect := func(v any) {
		refv := reflect.ValueOf(v)
		if v == nil {
			fmt.Println("v == nil")
			return
		}
		fmt.Printf("v.Type=%v(Kind:%v) .Value=%v .CanAddr=%t .CanInterface:%t .CanInterface:%t",
			refv.Type(), refv.Kind(), refv, refv.CanAddr(), refv.CanInterface(), refv.CanInterface())
		fmt.Printf(" .Implements(io.Reader):%t\n", refv.Type().Implements(typeofreader))
	}

	inspect(r)
	inspect(make(map[string]any))
	inspect(make(chan bool, 1))
	inspect(main)
	inspect(struct{ x, y int }{x: 35, y: 37})
	inspect(nopReader{})
	inspect(typeofreader)

	age := int64(26)
	name := "Liuxiang"

	field0 := reflect.StructField{
		Name:   "Age",
		Type:   reflect.TypeOf(age),
		Offset: uintptr(0),
		Index:  []int{0},
	}
	field1 := reflect.StructField{
		Name:   "Name",
		Type:   reflect.TypeOf(name),
		Tag:    reflect.StructTag(`json:"name"`),
		Offset: uintptr(1),
		Index:  []int{0, 1},
	}
	person := reflect.New(reflect.StructOf([]reflect.StructField{field0, field1}))
	person.Elem().Field(0).SetInt(age)
	person.Elem().Field(1).SetString(name)

	inspect(person)

	inspect(struct {
		Age  int    `json:"age"`
		Name string `json:"name"`
	}{Age: 26})

	inspect("")
}
