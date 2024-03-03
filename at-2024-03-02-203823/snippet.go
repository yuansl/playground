// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2024-03-02 20:38:23

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

	"github.com/qbox/net-deftones/util"
)

func main() {
	fp, err := os.Open("/tmp/loglinkctl.log")
	if err != nil {
		util.Fatal("os.Open:", err)
	}
	defer fp.Close()

	for reader := bufio.NewReader(fp); ; {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if !errors.Is(err, io.EOF) {
				util.Fatal("bufio.Reader.ReadBytes:", err)
			}
			break
		}
		line = bytes.TrimRight(line, "\n")

		findfield := func(name string) string {
			pos := bytes.Index(line, []byte(name))
			if pos == -1 {
				return ""
			}
			stop := bytes.Index(line[pos+len(name):], []byte(":"))
			field := line[pos+len(name) : pos+len(name)+stop]
			pos = bytes.Index(field, []byte(" "))
			return string(field[:pos])
		}

		if bytes.Contains(line, []byte("kodofs://")) {
			domain := findfield("Domain:")
			hour := findfield("Hour:")
			size := findfield("Traffic:")
			fmt.Printf("%s %s, size=%v\n", domain, hour, size)
		}
	}
}
