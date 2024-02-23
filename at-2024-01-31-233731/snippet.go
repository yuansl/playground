// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2024-01-31 23:37:31

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"bytes"
	"fmt"
)

func main() {
	s1 := "a"
	s2 := "a"
	buf := bytes.NewBuffer(make([]byte, len(s1)))
	var slice []byte

	fmt.Printf("len(slice)=%d, cap(slice)=%d\n", len(slice), cap(slice))

	buf.Write([]byte(s2))
	s3 := string(buf.Bytes())
	fmt.Println(s1 == s3, s1 == s2)
	fmt.Printf("s3='%v',s1='%s'\n", s3, s1)
}
