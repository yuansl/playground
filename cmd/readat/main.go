package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
)

const BUFSIZE = 16 << 10 // 16KiB

var (
	filename string
	c        int
	pprof    string
)

func init() {
	flag.StringVar(&filename, "f", "", "specify filename to read")
	flag.IntVar(&c, "c", runtime.NumCPU(), "concurrency")
	flag.StringVar(&pprof, "pprof", ":6060", "go tool pprof debug http addr")
}

func main() {
	flag.Parse()

	go func() {
		http.ListenAndServe(pprof, nil)
	}()

	var counter int64

	fp, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "os.Open failed: %v\n", err)
		os.Exit(1)
	}

	stat, err := fp.Stat()
	if err != nil {
		fmt.Fprintf(os.Stderr, "f.Stat failed: %v\n", err)
		os.Exit(2)
	}

	bufPool := sync.Pool{
		New: func() interface{} {
			atomic.AddInt64(&counter, +1)
			// fmt.Printf("#%d sync.Pool.New called again\n", atomic.LoadInt64(&counter))
			return make([]byte, BUFSIZE)
		},
	}

	cb := func(off int64, nread int, err error) {
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Fprintf(os.Stderr, "reader.ReadAt failed: %v\n", err)
			}
			return
		}
		// fmt.Printf("read at %d done\n", off)
	}

	var wg sync.WaitGroup
	var climit = make(chan struct{}, c)
	for off := int64(0); off < stat.Size(); off += BUFSIZE {
		climit <- struct{}{}

		wg.Add(1)
		go func(reader io.ReaderAt, off int64, callback func(off int64, nread int, err error)) {
			defer func() {
				<-climit
				wg.Done()
			}()

			var buf = bufPool.Get().([]byte)
			defer bufPool.Put(buf)

			n, err := reader.ReadAt(buf, off)
			callback(off, n, err)
		}(fp, off, cb)
	}
	wg.Wait()
}
