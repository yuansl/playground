// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-05-28 15:32:40

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

func fatal(v ...any) {
	fmt.Fprintln(os.Stderr, "fatal error: ", v)
	os.Exit(1)
}

const _BLOCK_SIZE = 1 << 20

func main() {
	fp, err := os.Open("/home/yuansl/Downloads/linuxmint-21.1-cinnamon-64bit.iso")
	if err != nil {
		fatal("os.Open:", err)
	}
	defer fp.Close()

	size := 0

	consume := func(buf []byte) {
		size += len(buf)
	}

	fstat, _ := fp.Stat()

	for off := int64(0); off < fstat.Size(); off += _BLOCK_SIZE {
		buf := [_BLOCK_SIZE]byte{}

		n, err := fp.ReadAt(buf[:], off)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				fatal("fp.ReadAt:", err)
			}
			if n > 0 {
				consume(buf[:n])
			}
			break
		}

		consume(buf[:n])
	}

	fmt.Printf("read %d bytes in total\n", size)
}
