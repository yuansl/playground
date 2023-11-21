// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-11-21 08:46:07

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"flag"
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"golang.org/x/sync/errgroup"
)

const _NR_TASKS = 10_000

var (
	_concurrency int
	_numberBench int
)

func parseCmdArgs() {
	flag.IntVar(&_concurrency, "c", runtime.NumCPU(), "concurrency")
	flag.IntVar(&_numberBench, "n", 10, "number of benchs")
	flag.Parse()
}

func f1() {
	var taskch = make(chan int, _concurrency)
	var done = make(chan bool)

	go func() {
		var egroup errgroup.Group

		for i := 0; i < _concurrency; i++ {
			egroup.Go(func() error {
				for num := range taskch {
					_ = num
					time.Sleep(10 * time.Microsecond)
				}
				return nil
			})
		}
		egroup.Wait()
		done <- true
	}()
	go func() {
		for i := 0; i < _NR_TASKS; i++ {
			taskch <- rand.Int()
		}
		close(taskch)
	}()
	<-done
}

func f2() {
	var taskch = make(chan int, _concurrency)
	var done = make(chan bool)

	go func() {
		var egroup errgroup.Group

		egroup.SetLimit(_concurrency)

		for num := range taskch {
			_num := num
			egroup.Go(func() error {
				_ = _num
				time.Sleep(10 * time.Microsecond)
				return nil
			})
		}
		egroup.Wait()
		done <- true
	}()
	go func() {
		for i := 0; i < _NR_TASKS; i++ {
			taskch <- rand.Int()
		}
		close(taskch)
	}()
	<-done
}

func bench(name string, f func()) {
	var elapsed time.Duration

	for i := 0; i < _numberBench; i++ {
		start := time.Now()
		f()
		elapsed += time.Since(start)
	}
	fmt.Printf("%s %d ns/op = %v\n", name, _numberBench, elapsed/time.Duration(_numberBench))
}

func main() {
	parseCmdArgs()

	var egroup errgroup.Group
	for _, it := range []struct {
		f    func()
		name string
	}{{name: "f2:one_goroutine_per_task", f: f2}, {name: "f1:n_goroutines", f: f1}} {
		_it := it
		egroup.Go(func() error { bench(_it.name, _it.f); return nil })
	}
	egroup.Wait()
}
