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

func signbit(x int) int {
	return ((^x << 59) >> 63)
}

func nextline(text string) {
	word := binary.BigEndian.Uint64([]byte(text))
	match := word ^ LINE_PATTERN

	fmt.Printf("word=%#x, pattern=%#x, match=%#x, mask1=%#x, (match-mask1)=%#x, %#x, %#08x\n", word, LINE_PATTERN, match, MASK1, (match - MASK1), (^match & MASK2), 0x04010203-0x03010101)

	pos := ((match - MASK1) & (^match & MASK2))

	fmt.Printf("pos=%#016X\n", pos)

	if newline := pos != 0; newline {
		fmt.Printf("OK, find new line\n")
	}
}

func main() {
	text := "Hamburg\n"

	nextline(text)

	for _, c := range text {
		fmt.Printf("%[1]c: %#02[1]x, %#08[1]b\n", c)
	}

	sign := signbit(3)
	fmt.Printf("sign=%d, 0x00-0x01=%#06x\n", sign, 0x010000-0x010101)
}
