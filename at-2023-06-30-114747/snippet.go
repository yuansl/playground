// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-06-30 11:47:47

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
)

func main() {
	var m sync.Map

	m.Store("a", 1)
	m.Store("b", 3)
	m.Store("c", 4)
	m.Store("d", 5)

	m.Range(func(key, value any) bool {
		if key == "d" {
			m.Delete("d")
		}
		if value.(int) == 1 {
			m.Delete(key)
		}
		m.Swap(key, 7)
		return true
	})

	m.Range(func(key, value any) bool {
		fmt.Println("key=", key, ", value=", value)
		return true
	})
}
