package main

import (
	"fmt"
	"os"
	"sort"

	"golang.org/x/exp/constraints"
)

func fatal(v ...any) {
	fmt.Fprintln(os.Stderr, "fatal error: ", v)
	os.Exit(1)
}

type numeric interface {
	constraints.Integer | constraints.Float
}

func AddArrayto[T numeric](a, b []T) []T {
	if len(a) != len(b) {
		panic("BUG: len(a)!=len(b)")
	}
	for i, x := range b {
		a[i] += x
	}
	return a
}

func peak95of[T numeric](nums []T) T {
	index := len(nums) / 20
	if index > 0 {
		index--
	}
	sort.Slice(nums, func(i int, j int) bool { return nums[i] > nums[j] })

	return nums[index]
}

func peakOf[T numeric](inputs []T) T {
	if len(inputs) == 0 {
		return T(0)
	}

	sort.Slice(inputs, func(i, j int) bool { return inputs[i] > inputs[j] })
	return inputs[0]
}

func sum[T numeric](inputs []T) (x T) {
	for _, in := range inputs {
		x += in
	}
	return x
}
