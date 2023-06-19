// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-03-27 08:20:14

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"flag"
	"fmt"
	"sync"
	"unsafe"
)

var usage = func() {
	println("hello1")
}

func CommmandLineUsage() {
	usage()
}

type noCopyable = sync.RWMutex

type Key struct {
	_ noCopyable
}

func main() {
	var uniq = make(map[Key]struct{})

	var key1 = Key{}
	var key2 = key1
	_ = key2

	_ = uniq

	var x = [0]int{}
	_ = x

	fmt.Println("sizeof(key1)=", unsafe.Sizeof(key1))

	flag.CommandLine.Usage = CommmandLineUsage

	usage = func() {
		println("hello2")
	}

	flag.CommandLine.Usage()

	usage = func() {
		println("hello3")
	}

	flag.CommandLine.Usage()
}
