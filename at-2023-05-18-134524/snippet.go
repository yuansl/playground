// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-05-18 13:45:24

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"fmt"
	"os"
	"reflect"
)

func main() {
	var r any = (*os.File)(nil)

	refv := reflect.ValueOf(r)

	fmt.Printf("refv.IsNil: %t r==nil: %t\n", refv.IsNil(), r == nil)
}
