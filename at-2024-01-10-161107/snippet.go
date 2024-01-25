// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2024-01-10 16:11:07

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"log"
	"sync"
	"sync/atomic"
)

type locker struct {
	lock atomic.Int32
}

func (l *locker) TryLock() bool { return l.lock.CompareAndSwap(0, 1) }
func (l *locker) Unlock() bool  { return l.lock.CompareAndSwap(1, 0) }

func createLocker() *locker { return &locker{} }

func main() {
	var lock = createLocker()
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			for attempts := 0; attempts < 10; attempts++ {
				if lock.TryLock() {
					log.Printf("#%d: got lock succeed at %d-th attempts\n", i, attempts)
					lock.Unlock()
				} else {
					log.Printf("#%d: got lock failed at %d-th attempts\n", i, attempts)
				}
			}
		}(i)
	}
	wg.Wait()
}
