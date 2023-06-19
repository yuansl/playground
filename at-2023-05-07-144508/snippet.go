// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-05-07 14:45:08

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	const NUM_SIZE = 10
	var nums [NUM_SIZE]int
	var rand1 = rand.New(rand.NewSource(time.Now().UnixNano()))
	_ = rand1

	for i := 0; i < len(nums); i++ {
		nums[i] = rand.Int() % 10000
	}
	fmt.Println("Results:", nums)
}
