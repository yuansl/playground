// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-05-27 22:26:51

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/yuansl/playground/at-2023-05-27-222651/repository"
	"github.com/yuansl/playground/at-2023-05-27-222651/repository/mysqlrepository"
)

const _BLOCK_SIZE = 1 << 20

type (
	Repository = repository.Repository
	File       = repository.File
)

func saveFile(filename string, repository Repository) {
	fp, err := os.Open(filename)
	if err != nil {
		fatal("os.Open:", err)
	}
	defer fp.Close()

	fstat, _ := fp.Stat()

	fmt.Println("file:", filename, " size:", fstat.Size())

	doSave := func(buf []byte, off int64) {
		fmt.Printf("Saving file %s: %d bytes at offset %d...\n", filename, len(buf), off)
		f := File{
			Name:    filepath.Base(fp.Name()),
			Size:    fstat.Size(),
			Offset:  off,
			Content: buf,
		}
		err = repository.Save([]*File{&f})
		if err != nil {
			fatal(err)
		}
	}
	for off := int64(0); off < fstat.Size(); off += _BLOCK_SIZE {
		buf := [_BLOCK_SIZE]byte{}

		n, err := fp.ReadAt(buf[:], off)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				fatal("fp.ReadAt:", err)
			}
			if n > 0 {
				// n > 0 on EOF
				doSave(buf[:n], off)
			}
			break
		}
		doSave(buf[:n], off)
	}
}

func readRepository(name string, off int64, repository Repository, w io.WriterAt) int64 {
	f := File{Name: name, Offset: off}
	err := repository.Load(&f)
	if err != nil {
		fmt.Printf("repository.Load error: %v\n", err)
		return -1
	}
	if f.Size == 0 {
		return -1
	}
	_, err = w.WriteAt(f.Content, off)
	if err != nil {
		fatal("fp.WriteAt:", err)
	}
	return f.Size
}

func loadFile(name string, repository Repository, w io.WriterAt) {
	filesize := readRepository(name, 0, repository, w)
	wg := sync.WaitGroup{}
	climit := make(chan struct{}, runtime.NumCPU())
	for off := int64(_BLOCK_SIZE); off < filesize; off += _BLOCK_SIZE {
		climit <- struct{}{}
		wg.Add(1)
		go func(off int64) {
			defer func() {
				<-climit
				wg.Done()
			}()

			fmt.Printf("loading file at offset: %d\n", off)
			readRepository(name, off, repository, w)
		}(off)
	}
	wg.Wait()
}

var filename string

func parseCmdArgs() {
	flag.StringVar(&filename, "f", "", "specify filename")
	flag.Parse()
}

func main() {
	parseCmdArgs()

	repository := mysqlrepository.NewRepository("yuansl:admin@(127.0.0.1:3306)/test?parseTime=true&loc=UTC&maxAllowedPacket=1073741824")

	saveFile(filename, repository)

	fp, err := os.OpenFile(filename+".2", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0664)
	if err != nil {
		fatal("os.OpenFile:", err)
	}
	defer fp.Close()

	loadFile(filepath.Base(filename), repository, fp)

	fmt.Printf("md5sum: %x\n", md5sumOf(fp.Name()))
}
