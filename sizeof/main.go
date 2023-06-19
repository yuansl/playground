package main

import (
	"fmt"
	"reflect"
	"runtime"
	"time"
	"unsafe"
)

func main() {
	var s []int

	sdesc := (*reflect.SliceHeader)(unsafe.Pointer(&s))
	fmt.Printf("s = %+v\n", sdesc)
}

func gc() {
	intptrs := make([]int, 0, 1e9)

	for i := 0; i < 10000; i++ {
		startAt := time.Now()
		runtime.GC()
		fmt.Printf("time elapsed: %v\n", time.Since(startAt))
	}

	runtime.KeepAlive(intptrs)
}
