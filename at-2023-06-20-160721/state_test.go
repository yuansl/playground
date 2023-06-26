package main

import (
	"math/rand"
	"testing"
)

func BenchmarkSwitch(b *testing.B) {
	var states = make([]State, 0, SIZE_MAX)

	for i := 0; i < SIZE_MAX; i++ {
		random := rand.Int()
		state := State(random % 5)

		if state > Received {
			state = Received
		}
		states = append(states, state)
	}

	b.Run("benchmark all-in-one-switch", func(b *testing.B) {
		var pstate = make(map[State]int)

		for _, state := range states {
			switch state {
			case Received:
				pstate[state]++
			case Connected:
				pstate[state]++
			case Caugth:
				pstate[state]++
			case Terminated:
				pstate[state]++
			default:
				pstate[state]++
			}
		}
	})

	b.Run("benchmark if-else-switch", func(b *testing.B) {
		var stateCounter = make(map[State]int)

		for _, state := range states {
			if state == Received {
				stateCounter[state]++
			} else {
				switch state {
				case Connected:
					stateCounter[state]++
				case Caugth:
					stateCounter[state]++
				case Terminated:
					stateCounter[state]++
				default:
					stateCounter[state]++
				}
			}
		}
	})
}
