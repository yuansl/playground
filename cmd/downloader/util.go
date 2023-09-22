package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

func fatal(v ...any) {
	formatted := func() bool {
		if len(v) == 0 {
			return false
		}
		if format, ok := v[0].(string); ok && strings.Contains(format, "%") {
			return true
		}
		return false
	}
	if formatted() {
		fmt.Fprintf(os.Stderr, "fatal error: "+v[0].(string), v[1:]...)
	} else {
		fmt.Fprintln(os.Stderr, "fatal error: ", v)
	}
	pc, file, line, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	fmt.Fprintf(os.Stderr, "%s:\n\t%s:%d\n", fn.Name(), file, line)
	os.Exit(1)
}

type Option interface{ apply(any) }

type OptionFn func(any)

func (fn OptionFn) apply(op any) { fn(op) }
