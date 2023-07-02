package main

import (
	"fmt"
	"os"
	"unicode"
)

func fatalf(format string, v ...any) {
	fmt.Fprintf(os.Stderr, "fatal error: "+format, v...)
	os.Exit(1)
}

func fatal(v ...any) {
	fmt.Fprintln(os.Stderr, "fatal error: ", v)
	os.Exit(1)
}

func skipSpace(s string) int {
	spaces := 0

	for i := 0; i < len(s) && unicode.IsSpace(rune(s[i])); i++ {
		spaces++
	}
	return spaces
}
