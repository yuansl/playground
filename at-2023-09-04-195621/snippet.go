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

	"github.com/qbox/net-deftones/util"
)

var fatal = util.Fatal

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
