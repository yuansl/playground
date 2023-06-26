// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-06-21 22:33:57

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

//go:generate stringer -type State
type State int

const (
	Connected State = iota
	Interrupted
	Received
	Terminated
	InvalidState
)

const SIZE = 10000

func process(states []State) float64 {
	stateCounter := make(map[State]int)

	for _, state := range states {
		switch state {
		case Received:
			stateCounter[state]++
		case Connected:
			stateCounter[state]++
		case Interrupted:
			stateCounter[state]++
		case Terminated:
			stateCounter[state]++
		default:
		}
	}
	fmt.Println("stateCounter=", stateCounter)

	return float64(stateCounter[Received]) / float64(stateCounter[Received]+stateCounter[Connected]+stateCounter[Terminated])
}

func main() {
	var states [SIZE]State

	for i := 0; i < SIZE; i++ {
		state := State(rand.Int() % int(InvalidState))

		if state > Received {
			state = Received
		}
		states[i] = state
	}

	fmt.Println("process: ", process(states[:]))
}
