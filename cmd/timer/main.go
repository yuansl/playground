package main

import (
	"fmt"
	"time"
)

type timer struct {
	C chan time.Time
}

func NewTimer(d time.Duration) timer {
	t := timer{C: make(chan time.Time)}
	go func() {
		defer close(t.C)
		time.Sleep(d)
		t.C <- time.Now()
	}()
	return t
}

func main() {
	t := NewTimer(10 * time.Second)
	t0 := time.NewTimer(9 * time.Second)

	select {
	case ts := <-t.C:
		fmt.Printf("timer timeouted at %v\n", ts)
	case ts := <-t0.C:
		fmt.Printf("time.Timer timeouted at %v\n", ts)
	}
}
