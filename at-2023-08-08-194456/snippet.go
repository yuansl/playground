// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-08-08 19:44:56

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import "errors"

type BucketUploader struct{}

var errNotImplemented = errors.New("not implemented")

func (*BucketUploader) UploadFile(filename string) error    { return errNotImplemented }
func (*BucketUploader) Download(key string) ([]byte, error) { return nil, errNotImplemented }

func main() {

}
