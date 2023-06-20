// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-06-20 16:07:21

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
)

const SIZE_MAX = 10000

//go:generate stringer -type State
type State int

const (
	Connected State = iota
	Idle
	Received
	Caugth
	Terminated
	Invalid = Terminated
)

func main() {
	var stateCounter = make(map[State]int)

	for i := 0; i < SIZE_MAX; i++ {
		random := rand.Int()
		state := State(random % 5)

		if state > Received {
			state = Received
		}
		stateCounter[state]++
	}

	for state, count := range stateCounter {
		fmt.Printf("state=%s, count=%d\n", state, count)
	}
}
