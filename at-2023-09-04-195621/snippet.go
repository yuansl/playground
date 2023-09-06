// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-09-04 19:56:21

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
)

func fatal(v ...any) {
	formatted := func() bool {
		if len(v) == 0 {
			return false
		}
		if format, ok := v[0].(string); ok && strings.Contains(format, "%") {
			return true
		}
		return false
	}
	if formatted() {
		fmt.Fprintf(os.Stderr, "fatal error: "+v[0].(string), v[1:]...)
	} else {
		fmt.Fprintln(os.Stderr, "fatal error: ", v)
	}
	pc, file, line, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	fmt.Fprintf(os.Stderr, "%s:\n\t%s:%d\n", fn.Name(), file, line)
	os.Exit(1)
}

func main() {
	fp, err := os.Open("/tmp/some.tar.gz")
	if err != nil {
		fatal("os.Open:", err)
	}
	defer fp.Close()

	var buf bytes.Buffer

	gz, err := gzip.NewReader(fp)
	if err != nil {
		fatal("gzip.NewReader:", err)
	}

	if _, err := io.Copy(&buf, gz); err != nil {
		fatal("io.Copy:", err)
	}
	if err := gz.Close(); err != nil {
		fatal("gzip.Close:", err)
	}

	md5sum := md5.Sum(buf.Bytes())

	md5sumtext := hex.EncodeToString(md5sum[:])

	fmt.Println("Results:", md5sumtext)
}
