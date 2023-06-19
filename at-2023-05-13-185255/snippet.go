// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-05-13 18:52:55

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"fmt"
	"io"
	"runtime"

	"github.com/pkg/errors"
)

func stackTrace(depth int) {
	pc, file, line, ok := runtime.Caller(0 + depth)
	if ok {
		fn := runtime.FuncForPC(pc)
		fmt.Printf("	%s:%d:%s\n", file, line, fn.Name())
	}
}

type xError struct{}

func (*xError) Error() string {
	stackTrace(+2)
	return "xError: io.EOR"
}

func (*xError) Cause() error { stackTrace(+3); return fmt.Errorf("%w -> xError: io error", io.EOF) }

func main() {
	println("cause:" + errors.Cause(&xError{}).Error())
	e := &xError{}
	fmt.Printf("xError: %v\n", e.Error())
	// e := errors.New("hello")

	// fmt.Println("error:" + e.Error())
	// btrace, ok := e.(interface{ StackTrace() errors.StackTrace })
	// if ok {
	// 	fmt.Printf("stack trace: %+v\n", btrace.StackTrace())
	// }
}
