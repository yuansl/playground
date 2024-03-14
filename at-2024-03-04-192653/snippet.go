// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2024-03-04 19:26:53

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"fmt"
	"sync"
	"time"
)

type Operation struct {
	Start, End time.Time
	Op         string
}

type Watcher struct {
	mu  sync.RWMutex
	ops []Operation
}

func (w *Watcher) Add(op string, start time.Time) {
	o := Operation{Op: op, Start: start, End: time.Now()}
	w.mu.Lock()
	w.ops = append(w.ops, o)
	w.mu.Unlock()
}

func (w *Watcher) Stop() {
	w.mu.RLock()
	defer w.mu.RUnlock()
	for _, op := range w.ops {
		fmt.Printf("'%s', started at %s, stopped at %s, time elapsed %v\n", op.Op, op.Start, op.End, op.End.Sub(op.Start))
	}
	w.ops = w.ops[:0]
}

func main() {
	var watcher Watcher

	func() { defer watcher.Add("func1", time.Now()); time.Sleep(2 * time.Millisecond) }()
	func() { defer watcher.Add("func2", time.Now()); time.Sleep(5 * time.Millisecond) }()
	func() { defer watcher.Add("func3", time.Now()); time.Sleep(10 * time.Millisecond) }()
	func() { defer watcher.Add("func4", time.Now()); time.Sleep(10 * time.Millisecond) }()

	watcher.Stop()
}
