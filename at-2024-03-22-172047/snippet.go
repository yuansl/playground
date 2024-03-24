// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2024-03-22 17:20:47

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/golang/snappy"
	"github.com/yuansl/playground/util"
)

func main() {
	fp, err := os.Open("a.snappy")
	if err != nil {
		util.Fatal(err)
	}
	defer fp.Close()
	for reader := bufio.NewReader(snappy.NewReader(fp)); ; {
		var buf [4096]byte
		n, err := reader.Read(buf[:])
		if err != nil {
			switch {
			case errors.Is(err, io.EOF):
			default:
				util.Fatal("reader.Read:", err)
			}
			break
		}
		fmt.Printf("line = `%s`\n", buf[:n])
	}
}
