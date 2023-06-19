package main

//go:generate protoc -I ./proto --go_out=./proto --go_opt=paths=source_relative --grpc_out=./proto --grpc_opt=paths=source_relative ./proto/some.proto

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
)

type Misc struct {
	CdnName string `json:"cdn_name"`
}

type Traffic struct {
	Misc
	Domain  string  `json:"domain_name"`
	Data    []int64 `json:"points"`
	Private bool
	Extra   struct {
		Cdn string `json:"cdn_provider"`
	} `json:"extra_info"`
	Props map[string]any
}

func (traffic *Traffic) Read(p []byte) (int, error) {
	b := bytes.NewBuffer(p)
	return b.WriteString(fmt.Sprintf("Traffic{CdnName:%s,Domain=%q,Data=%b,Private=%t,Extra=%v,Props=%v}", traffic.CdnName, traffic.Domain, traffic.Data, traffic.Private, traffic.Extra, traffic.Props))
}

func main() {
	traffic := Traffic{
		Misc:   Misc{CdnName: "aliyun"},
		Domain: "www.example.com",
		Data:   []int64{1, 2, 3, 4, 5},
		Props: map[string]any{
			"City":    "hangzhou",
			"Country": "China",
		},
	}
	traffic.Extra.Cdn = "aliyun cdn service"

	type reader interface{ Read(p []byte) (int, error) }

	var trafficReader reader = &traffic

	inspect(trafficReader.(interface{ io.Reader }))

	http.StripPrefix("/statd/", http.FileServer(http.Dir("/var/www")))
}

func inspect(v any) {
	refv := reflect.ValueOf(v)

	switch kind := refv.Kind(); kind {
	case reflect.Pointer, reflect.Interface:
		fmt.Printf("Kind: %v\n", kind)
		inspect(refv.Elem().Interface())
	case reflect.Struct:
		inspectStruct(v)
	case reflect.Array, reflect.Slice, reflect.Chan, reflect.String:
		fmt.Printf(" type=%q: #%d(%#v)\n", refv.Type(), refv.Len(), refv.Interface())
		if kind != reflect.String {
			for i := 0; i < refv.Len(); i++ {
				inspect(refv.Index(i).Interface())
			}
		}
	case reflect.Map:
		for i, key := range refv.MapKeys() {
			value := refv.MapIndex(key)
			fmt.Printf("Map #%d: (%v -> %v)\n", i, key, value)
		}
	case reflect.Bool:
		fmt.Printf(" type=%q: value: %t\n", refv.Type(), refv.Interface())
	case reflect.Float32, reflect.Float64:
		fmt.Printf(" type=%q: value: %.2f\n", refv.Type(), refv.Interface())
	case reflect.Int64, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		fmt.Printf(" type=%q: %d\n", refv.Type(), refv.Interface())
	default:
		fmt.Printf("Unknown type of identifier: %v, refv=%v\n", kind, refv)
	}
}

func inspectString(v any) string {
	refv := reflect.ValueOf(v)
	if refv.Kind() != reflect.String {
		return ""
	}
	return fmt.Sprintf("string: identifier=%q, value=%q", refv.Type(), refv)
}

func inspectStruct(v any) {
	refv := reflect.ValueOf(v)

	if refv.Kind() != reflect.Struct {
		return
	}
	fmt.Printf("Name=%q, (kind)%v, assignable: %t fields:\n", refv.Type(), refv.Kind(), refv.Type().AssignableTo(reflect.TypeOf((*Traffic)(nil)).Elem()))

	refv3_45 := reflect.ValueOf(3.45)

	fmt.Printf("3.45 CovertibleTo 3: %t, %d\n", refv3_45.Type().ConvertibleTo(reflect.TypeOf((*int)(nil)).Elem()), refv3_45.Convert(reflect.TypeOf((*int)(nil)).Elem()))

	refvints := reflect.ValueOf([]int{1, 3, 5, 7, 9})

	fmt.Printf("&ints[1]=%p\n", refvints.Index(1).Addr().Interface())

	refvints.Index(1).Set(reflect.ValueOf(10))

	fmt.Printf("revints=%d\n", refvints.Index(1).Int())
	var x = 3.45

	refvx := reflect.ValueOf(&x)

	fmt.Printf("revx.CanSet: %t\n", refvx.Elem().CanSet())

	refvx.Elem().Set(reflect.ValueOf(3.88))

	fmt.Printf("revx.CanSet: %t, after set: %f\n", refvx.Elem().CanSet(), refvx.Elem().Interface())
	for i := 0; i < refv.NumField(); i++ {
		fieldv := refv.Field(i)
		fmt.Printf("fieldv.Type=%v\n", fieldv.Type())
		_type := refv.Type().Field(i)

		fmt.Printf("  #%d: Name=%q,Offset=%d,Tag=%v", i, _type.Name, _type.Offset, lookup(string(_type.Tag), "json"))
		if _type.Anonymous {
			fmt.Printf(",Embedded\n")
			inspect(fieldv.Interface())
			continue
		}
		inspect(fieldv.Interface())
		fmt.Println()
	}
}

func lookup(tag, name string) string {
	var tagvalue []byte
	idx := strings.Index(tag, name)
	if idx == -1 {
		return ""
	}

	if idx < len(tag) && tag[idx] == ':' {
		idx++ // skip ':'
	}
	for idx < len(tag) && tag[idx] != '"' {
		idx++
	}
	if idx < len(tag) && tag[idx] == '"' {
		idx++ // skip '"'
	}
	for ; idx < len(tag) && tag[idx] != '"'; idx++ {
		tagvalue = append(tagvalue, tag[idx])
	}
	return string(tagvalue)
}
