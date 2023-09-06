// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-09-04 15:05:32

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime"
	"strings"
)

func fatal(v ...any) {
	pc, file, line, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)

	fmt.Fprintln(os.Stderr, "fatal error: ", v)
	fmt.Fprintf(os.Stderr, "%s:\n\t%s:%d\n", fn.Name(), file, line)
	os.Exit(1)
}

func runHttpServer(addr string) error {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		raw, _ := httputil.DumpRequest(req, false)
		fmt.Printf("Request(raw): %q\n", raw)
		if strings.Contains(req.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(req.Body)
			if err != nil {
				fatal("gzip.NewReader:", err)
			}

			if strings.Contains(req.Header.Get("Content-Type"), "application/json") {
				data, _ := io.ReadAll(gz)

				var v map[string]any

				if err := json.Unmarshal(data, &v); err != nil {
					fatal("json.NewDecoder:", err, "raw json data:", string(data))
				}
				fmt.Printf("Read from request: '%+v'\n", v)
				return
			}

			for {
				var buf [4096]byte

				n, err := gz.Read(buf[:])
				if err != nil {
					if n > 0 {
						fmt.Printf("Read from request (gz stream): %q\n", buf[:n])
					}
					if !errors.Is(err, io.EOF) {
						fatal("gz.Read:", err)
					}
					break
				}
				fmt.Printf("Read from request (gz stream): %q\n", buf[:n])
			}
		}
	})
	return http.ListenAndServe(addr, nil)
}

func main() {
	runHttpServer(":8080")
}
