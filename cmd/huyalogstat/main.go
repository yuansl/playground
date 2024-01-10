// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2024-01-09 15:44:25

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
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/yuansl/playground/util"
)

const BUFSIZ = 16 << 10

type Log struct {
	Timestamp time.Time
	Domain    string
	Url       string
	BytesSent int64
}

func statLogLine(r io.Reader) {
	var uniq = make(map[string]struct{})
	var flux int64

	for r := bufio.NewReaderSize(r, BUFSIZ); ; {
		line, err := r.ReadBytes('\n')
		if err != nil {
			if !errors.Is(err, io.EOF) {
				util.Fatal("bufio.ReadBytes: ", err)
			}
			break
		}
		line = bytes.TrimSpace(line)

		if _, exists := uniq[string(line)]; exists {
			continue
		} else {
			uniq[string(line)] = struct{}{}
		}

		fields := bytes.Split(line, []byte("#_#"))
		if len(fields) <= 15 {
			fmt.Fprintf(os.Stderr, "WARNING: bad format in line: '%s'\n", line)
			continue
		}
		bytes, err := strconv.Atoi(string(fields[14]))
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: bad value in fields[15]: '%s'\n", fields[14])
			continue
		}
		flux += int64(bytes)
	}

	fmt.Printf("flux = %d\n", flux)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <filename>(in gzip format)", os.Args[0])
		os.Exit(1)
	}
	fp, err := os.Open(os.Args[1])
	if err != nil {
		util.Fatal(err)
	}
	gz, err := gzip.NewReader(fp)
	if err != nil {
		util.Fatal(err)
	}
	defer gz.Close()

	statLogLine(gz)
}
