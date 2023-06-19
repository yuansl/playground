// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-03-25 21:06:41

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"encoding"
	"encoding/json"
	"fmt"
	"os"
)

type Person struct {
	Age     int
	Name    string
	*Person `json:"Person"`
}

func fatal(v ...any) {
	fmt.Fprintf(os.Stderr, "fatal error: %v\n", v...)
	os.Exit(1)
}

func main() {
	var john = Person{
		Name: "john",
		Age:  26,
	}
	john.Person = &john

	data, err := json.Marshal(john)
	if err != nil {
		fatal(err)
	}
	fmt.Printf("data=%q\n", data)

	var _ encoding.BinaryMarshaler
}
