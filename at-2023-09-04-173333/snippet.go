// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-09-04 17:33:33

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"archive/tar"
	"fmt"
	"os"
	"runtime"
	"time"
)

func fatal(v ...any) {
	pc, file, line, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)

	fmt.Fprintln(os.Stderr, "fatal error: ", v)
	fmt.Fprintf(os.Stderr, "%s:\n\t%s:%d\n", fn.Name(), file, line)
	os.Exit(1)
}

func main() {
	fp, err := os.OpenFile("/tmp/some.tar", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0664)
	if err != nil {
		fatal("os.OpenFile:", err)
	}
	tarw := tar.NewWriter(fp)

	{
		content := []byte("hello, world")
		tarw.WriteHeader(&tar.Header{
			Name:       "a",
			Size:       int64(len(content)),
			Mode:       0664,
			ModTime:    time.Now(),
			AccessTime: time.Now(),
			Uid:        os.Geteuid(),
			Gid:        os.Getgid(),
			Uname:      os.Getenv("USER"),
			Gname:      os.Getenv("USER"),
		})
		tarw.Write(content)
	}

	{
		content := []byte("hello, world, this is file b")
		tarw.WriteHeader(&tar.Header{
			Name:       "b",
			Size:       int64(len(content)),
			Mode:       0664,
			ModTime:    time.Now(),
			AccessTime: time.Now(),
			Uid:        os.Geteuid(),
			Gid:        os.Getgid(),
			Uname:      os.Getenv("USER"),
			Gname:      os.Getenv("USER"),
		})
		tarw.Write(content)
		if err = tarw.Flush(); err != nil {
			fatal(err)
		}
	}

	if err = tarw.Close(); err != nil {
		fatal(err)
	}
}
