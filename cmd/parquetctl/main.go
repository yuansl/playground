// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-07-14 13:05:34

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"flag"
)

const BUFSIZE = 16 << 10

var _options struct {
	filename string
}

func parseCmdOptions() {
	flag.StringVar(&_options.filename, "f", "", "specify parquet filename")
	flag.Parse()
}

func main() {
	parseCmdOptions()

	// mergeParquetFiles(strings.Split(_options.filename, ","))

	InspectParquetFile(_options.filename)
}
