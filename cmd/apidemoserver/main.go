// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-10-15 12:10:31

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"context"

	"github.com/yuansl/playground/util"
)

func main() {
	conf := initializeConfig()
	srv := NewServer(conf)
	defer srv.Close()

	ctx := util.InitSignalHandler(context.TODO())

	if err := srv.Run(ctx); err != nil {
		util.Fatal(err)
	}
}
