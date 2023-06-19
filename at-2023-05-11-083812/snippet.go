// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-05-11 08:38:12

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
)

func maxOf(nums []int) (max int) {
	if len(nums) == 0 {
		return
	}
	max = nums[0]
	for i := 1; i < len(nums); i++ {
		if nums[i] > max {
			max = nums[i]
		}
	}
	return max
}

func countSort(nums []int) {
	countslot := make([]int, maxOf(nums)+1)
	nums2 := make([]int, len(nums))

	for _, num := range nums {
		countslot[num] += 1
	}
	for i := 1; i < len(countslot); i++ {
		countslot[i] += countslot[i-1]
	}
	for _, num := range nums {
		pos := countslot[num] - 1
		nums2[pos] = num
		countslot[num]--
	}
	for i, v := range nums2 {
		nums[i] = v
	}
}

type Map[K comparable, V any] struct {
	underlying map[K]V
}

func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{underlying: make(map[K]V)}
}

func (m *Map[K, V]) Put(key K, val V) {
	m.underlying[key] = val
}

type someReader struct {
	io.Reader
}

func (x *someReader) Read(p []byte) (int, error) {
	return 0, io.EOF
}

type fReader func([]byte) (int, error)

func (fn fReader) Read(b []byte) (int, error) {
	return fn(b)
}

const _NR_NUMS = 20

func foo3[T interface{ ~int | ~int8 | ~int16 }](a T, b T) {
	_ = a == b
}

func main() {
	nums := [_NR_NUMS]int{}

	for i := 0; i < _NR_NUMS; i++ {
		nums[i] = rand.Int() % 50
	}
	fmt.Println("before count-sort:", nums)
	countSort(nums[:])
	fmt.Println("after count-sort:", nums)

	readerMap := NewMap[io.Reader, any]()
	r1 := &bytes.Reader{}
	r2 := &bytes.Buffer{}
	// var r3, r5 fReader
	r4 := &someReader{}
	readerMap.Put(r1, 1)
	readerMap.Put(r2, 1)
	readerMap.Put(r4, 1)

	// _ = io.Reader(r3) == io.Reader(r5)

	type some struct {
		r fReader
	}
	// var r6, r7 some

	// var r8, r9 [0]fReader

	// fmt.Println("r1 == r2: ", r6 == r7, r8 == r9)
}
