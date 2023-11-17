package main

import "fmt"

func assert(expression bool, msg ...any) {
	if !expression {
		panic(fmt.Sprint("BUG:", msg))
	}
}
