// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-12-03 11:06:21

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"fmt"
)

func main() {
	fmt.Println("Results:", (1969/4 - 1969/100 + 1969/400))

	leapYearCounter := 0
	for year := 1; year <= 1970; year++ {
		if (year%4 == 0 && year%100 != 0) || year%400 == 0 {
			leapYearCounter++
		}
	}
	fmt.Printf("There are %d leap years since 1,1,1 UTC until 1970,1,1 UTC\n", leapYearCounter)
}
