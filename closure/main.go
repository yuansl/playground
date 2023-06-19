package main

import (
	"flag"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
)

const NUM_MAX = 100_0000

var concurrency int

func parseOptions() {
	flag.IntVar(&concurrency, "c", runtime.NumCPU(), "specify the concurrency level")
	flag.Parse()
}

func main() {
	parseOptions()

	var ch1 = make(chan []int, concurrency)
	var ch2 = make(chan int, concurrency)

	go func() {
		defer close(ch2)

		var wg sync.WaitGroup
		wg.Add(concurrency)
		for i := 0; i < concurrency; i++ {
			go func() {
				defer wg.Done()
				for vals := range ch1 {
					var x int
					for _, val := range vals {
						x += val
					}
					ch2 <- x
				}
			}()
		}
		wg.Wait()
	}()

	go func() {
		defer close(ch1)
		for i := 0; i < NUM_MAX; i++ {
			var vals [NUM_MAX / 20]int
			for i := 0; i < len(vals); i++ {
				vals[i] = rand.Int() % 1000000
			}
			ch1 <- vals[:]
		}
	}()

	var sum int
	for i := range ch2 {
		sum += i
	}
	fmt.Printf("sum=%d\n", sum)
}
