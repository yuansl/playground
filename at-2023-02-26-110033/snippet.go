// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-02-26 11:00:33

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"errors"
)

func main() {
	defer func() {
		defer func() {
			e := recover()
			println("e=", e)
			println("f1:", e.(error).Error())
		}()

		if e := recover(); e != nil {
			println("f2:", e == nil)
			panic(e) // rethrow
		}

	}()

	panic(errors.New("Oops"))
}
