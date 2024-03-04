// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2024-03-03 16:55:12

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"encoding/binary"
	"fmt"
)

const LINE_PATTERN = 0x0A0A0A0A0A0A0A0A
const MASK1 = 0x0101010101010101
const MASK2 = 0x8080808080808080

func signiess(x int) int {
	return ((^x << 59) >> 63) & 0xFF
}

func main() {
	text := "Hamburg;129.0\n"
	word := binary.LittleEndian.Uint64([]byte(text))
	match := word ^ LINE_PATTERN

	var _2 [2]byte

	fmt.Printf("12 ^ 10 = %d, '%c', -1=%#08b\n", 12^10, 0x0A, _2)

	fmt.Printf("world=%#x\n", (match-MASK1)&(^match&MASK2))

	for _, c := range text {
		fmt.Printf("%[1]c: %#02[1]x, %#08[1]b\n", c)
	}

	s := signiess(-3)
	fmt.Printf("s=%d\n", s)

}
