// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-09-23 23:49:53

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

func main() {
	var f func()
	for i := 0; i < 10; i++ {
		if i == 0 {
			f = func() { print(i) }
		}
		f()
	}
}
