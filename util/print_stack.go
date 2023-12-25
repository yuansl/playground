package util

import (
	"fmt"
	"runtime"
)

func PrintStackTrace() {
	var (
		pc     [128]uintptr
		n      = runtime.Callers(2, pc[:])
		frames = runtime.CallersFrames(pc[:n])
	)
	for {
		frame, next := frames.Next()
		fmt.Printf(" @%s:%s:%d\n", frame.File, frame.Function, frame.Line)
		if !next {
			break
		}
	}
}
