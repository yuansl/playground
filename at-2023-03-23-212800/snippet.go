// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-03-23 21:28:00

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
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

var filename string

func main() {
	flag.StringVar(&filename, "filename", "", "file to be collected")
	flag.Parse()
	fp, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	reader := bufio.NewReader(fp)

	count := 0
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if !errors.Is(err, io.EOF) {
				log.Fatal(err)
			}
			break
		}
		_ = line
		count++
	}
	fmt.Printf("Results: read %d lines in total\n", count)
}
