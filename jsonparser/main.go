// -*- mode:go-playground -*-
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	KiB = 1024
	MiB = 1024 * KiB
)

func jsonDecodeString() {
	rawjson := `{"name": "lio", "age": 26, "female": false, "extra": null, "points": [1,2,3,4,{"attr": "address", "location": "beijing"}, null, null, null, "what are you doing man? ", "hello", "\u0013ffff"]}`

	s := parseJson([]byte(rawjson))
	fmt.Printf("parsed json: %v\n", s)
}

func jsonDecodeFile() {
	fp, err := os.Open(filepath.Join(os.Getenv("HOME"), ".cache/mintinstall/reviews.json"))
	if err != nil {
		fatal("os.Open error:", err)
	}
	defer fp.Close()

	data, err := io.ReadAll(io.LimitReader(fp, 30*MiB))
	if err != nil {
		fatal("io.ReadAll error:", err)
	}

	jsonobj := parseJson(data)

	data, _ = json.MarshalIndent(jsonobj, "", " ")
	fmt.Printf("data=`%s`\n", data)
}

func main() {
	jsonDecodeFile()
}
