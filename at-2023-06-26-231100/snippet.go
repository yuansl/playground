// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-06-26 23:11:00

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
)

type RegionPair map[string]string

var ContinentAsRegions = map[string]string{
	"America": "AM",
	"Europe":  "em",
	"Africa":  "AF",
}

func _ContinentAsRegions() RegionPair {
	return map[string]string{
		"America": "AM",
		"Europe":  "em",
		"Africa":  "AF",
	}
}

const multiline = `
This is a long text
that spans multiple lines
and then ends.`

func fatal(v ...any) {
	fmt.Fprintln(os.Stderr, "fatal error: ", v)
	os.Exit(1)
}

type lineReader struct {
	*bufio.Reader
	eof bool
}

func (r *lineReader) Readline() ([]byte, error) {
	if r.eof {
		return nil, io.EOF
	}

	line, err := r.ReadBytes('\n')
	if err != nil {
		if !errors.Is(err, io.EOF) {
			fatal("r.ReadBytes:", err)
		}

		if len(line) > 0 {
			r.eof = true
			return line, nil
		}
	}
	return line, nil
}

func NewLineReader(r io.Reader) *lineReader {
	return &lineReader{Reader: bufio.NewReader(r)}
}

func main() {
	for r := NewLineReader(bytes.NewReader([]byte(multiline))); ; {
		line, err := r.Readline()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				fatal("r.ReadBytes:", err)
			}
			break
		}

		fmt.Printf("line=%q\n", line)
	}
}
