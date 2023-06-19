package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"time"
)

var n int

func init() {
	rand.Seed(time.Now().Unix())
}

func parseOptions() {
	flag.IntVar(&n, "n", 1000_000, "specify count of the tests")
	flag.Parse()
}

func main() {
	parseOptions()

	go func() {
		http.ListenAndServe(":6060", nil)
	}()

	// safeMap := safemap.NewSafeMap[string, int]()
	var syncMap sync.Map

	var wg sync.WaitGroup

	for i := 0; i < n; i++ {
		// safeMap.Set(fmt.Sprint(i), rand.Int())
		syncMap.Store(fmt.Sprint(i), rand.Int())
	}

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// safeMap.Get(fmt.Sprint(i))
			syncMap.Load(fmt.Sprint(i))
		}(i)
	}
	wg.Wait()
}
