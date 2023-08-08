package main

import (
	"fmt"
	"os"

	"golang.org/x/exp/constraints"
)

func max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func fatal(v ...any) {
	fmt.Fprintln(os.Stderr, "fatal error: ", v)
	os.Exit(1)
}
