// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-05-18 11:26:58

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"fmt"
	"playground/at-2023-05-18-112658/logger"
	"reflect"
)

func print(log logger.Logger) {
	log.Errorf("hello")
	log.Infof("hello")
	log.Warnf("hello")
	log.Debugf("hello")
}

func main() {
	log := logger.New(logger.LDebug)

	print(log)

	x := reflect.ValueOf(log).Type().Implements(reflect.TypeOf((*logger.Logger)(nil)).Elem())
	fmt.Printf("log implements Logger interface: %t\n", x)

	fmt.Println("reflect(log).NumMethod=", reflect.ValueOf(log).NumMethod(),
		" reflect(&log).NumMethod=", reflect.ValueOf(log).NumMethod())

	reftype := reflect.TypeOf(log)
	for i := 0; i < reftype.NumMethod(); i++ {
		fmt.Printf("method #%d: %v\n", i, reftype.Method(i).Name)
	}
}
