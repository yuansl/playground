// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-08-07 14:53:02

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

func fatal(v ...any) {
	fmt.Fprintln(os.Stderr, "fatal error: ", v)
	os.Exit(1)
}

func main() {
	raw, err := os.ReadFile("./some.log")
	if err != nil {
		fatal("os.ReadFile:", err)
	}

	fields := bytes.Split(raw, []byte("|#|"))

	if len(fields) >= 65 {
		data := fields[64]

		if i := bytes.Index(data, []byte("options=")); i >= 0 {
			data = bytes.TrimSpace(data[len("options="):])

			fmt.Printf("data='%s'\n", data)

			unquoted, err := strconv.Unquote(`"` + string(data) + `"`)
			if err != nil {
				fatal("strconv.Unquote error:", err)
			}

			var v map[string]string

			fmt.Printf("data='%q'\n", unquoted)

			if err = json.Unmarshal([]byte(unquoted), &v); err != nil {
				fatal("json.Unmarshal error:", err)
			}
			fmt.Printf("v=%v\n", v)
		}
	}
}
