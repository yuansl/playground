// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2024-03-07 14:24:27

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"container/heap"
	"fmt"
)

type Ints []int

// Len implements heap.Interface.
func (ints *Ints) Len() int {
	return len(*ints)
}

// Less implements heap.Interface.
func (ints *Ints) Less(i int, j int) bool {
	return (*ints)[i] > (*ints)[j]
}

// Pop implements heap.Interface.
func (ints *Ints) Pop() any {
	if len(*ints) == 0 {
		return nil
	}
	n := ints.Len()
	x := (*ints)[n-1]
	*ints = (*ints)[:n-1]

	return x
}

// Push implements heap.Interface.
func (ints *Ints) Push(x any) {
	*ints = append(*ints, x.(int))
}

// Swap implements heap.Interface.
func (ints *Ints) Swap(i int, j int) {
	(*ints)[i], (*ints)[j] = (*ints)[j], (*ints)[i]
}

var _ heap.Interface = (*Ints)(nil)

func main() {
	nums := Ints([]int{1, 3, 24, 8, 0, -1})

	heap.Init(&nums)

	for nums.Len() > 0 {
		tmp := heap.Pop(&nums)
		fmt.Printf("x =%v\n", tmp)
	}
}
