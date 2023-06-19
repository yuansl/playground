// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-06-03 13:39:30

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import "testing"

type some struct {
	Name         string `json:"name"`
	Age          int    `json:"age"`
	ProvinceCity string `json:"province_city"`
}

func TestFoo2(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Foo2()
		})
	}
}
