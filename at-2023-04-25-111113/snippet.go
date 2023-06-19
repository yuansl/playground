// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-04-25 11:11:13

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
	"math/rand"
	"os"
)

func fatal(v ...any) {
	fmt.Fprintln(os.Stderr, "fatal error: ", v)
	os.Exit(1)
}

func rewrite(w io.Writer, r io.Reader) error {
	reader := bufio.NewReader(r)

	for count := 0; ; count++ {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if !errors.Is(err, io.EOF) {
				fatal(err)
			}
			break
		}
		if count%2 == 0 {
			tmp := make([]byte, len(line))
			copy(tmp, line)
			bytes.TrimSpace(tmp)
			fields := bytes.Fields(tmp)
			if len(fields) < 2 {
				continue
			}

			line = line[:0]
			line = append(line, fields[0]...)
			for i := 0; i < rand.Int()%10; i++ {
				line = append(line, '\000')
			}
			line = append(line, fields[1]...)
			w.Write([]byte{'\n'})
		}
		w.Write(line)
	}
	return nil
}

func main() {
	filename := os.Args[1]
	fp, err := os.Open(filename)
	if err != nil {
		fatal("os.Open error:", err)
	}
	defer fp.Close()

	tmpfile, err := os.OpenFile(filename+".tmp",
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fatal("os.CreateTemp error:", err)
	}
	defer tmpfile.Close()

	err = rewrite(tmpfile, fp)
	if err != nil {
		fatal("rewrite error:", err)
	}

	err = os.Rename(tmpfile.Name(), filename)
	if err != nil {
		fatal("os.Rename:", err)
	}
}
