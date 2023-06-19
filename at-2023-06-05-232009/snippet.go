// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-06-05 23:20:09

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"context"

	"playground/at-2023-06-05-232009/repository"
	"playground/at-2023-06-05-232009/repository/dbrepository"
)

func run(ctx context.Context, repo repository.Repository) error {
	repo.ListDomains(ctx)

	return nil
}

func main() {
	repo := dbrepository.NewRepository()

	run(context.TODO(), repo)
}
