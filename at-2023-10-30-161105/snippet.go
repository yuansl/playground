// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-10-30 16:11:05

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"bytes"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/yuansl/playground/util"
)

var _bytesPoolNewCounter atomic.Int64

var _bytesPool = sync.Pool{
	New: func() any {
		_bytesPoolNewCounter.Add(+1)
		return make([]byte, 8<<10)
	},
}

func main() {
	for i := 0; i < 100000; i++ {
		func(i int) {
			buf := _bytesPool.Get().([]byte)
			defer func() { _bytesPool.Put(buf) }()

			b := bytes.NewBuffer(buf)
			if _, err := fmt.Fprintf(b, "hello, this is #%d calling\n", i); err != nil {
				util.Fatal(err)
			}
		}(i)
	}

	fmt.Printf("_bytesPoolNewCounter=%d\n", _bytesPoolNewCounter.Load())
}
