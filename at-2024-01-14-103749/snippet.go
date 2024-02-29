// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2024-01-14 10:37:49

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/qbox/net-deftones/util"
	"golang.org/x/sys/unix"
)

var _options struct {
	filename string
	ioengine string
}

func parseCmdOptions() {
	flag.StringVar(&_options.filename, "f", "", "specify filename")
	flag.StringVar(&_options.ioengine, "ioengine", "go", "ioengine. one of 'go'|'pread'|'read'")
	flag.Parse()
}

const READER_BUFSIZE = 16 << 10

func PreadInParallel(filename string) {
	fp, err := os.Open(filename)
	if err != nil {
		util.Fatal(err)
	}
	defer fp.Close()

	stat, _ := fp.Stat()

	start := time.Now()

	var nbytes atomic.Int64

	var wg sync.WaitGroup

	for off := int64(0); off < stat.Size(); off += READER_BUFSIZE {
		wg.Add(1)
		go func(off int64) {
			defer wg.Done()
			var buf [READER_BUFSIZE]byte

			n, err := unix.Pread(int(fp.Fd()), buf[:], off)
			if err != nil {
				if !errors.Is(err, io.EOF) {
					util.Fatal("Pread(fd=%d,offset=%d): %v\n", int(fp.Fd()), off, err)
				}
				return
			}
			nbytes.Add(int64(n))
		}(off)
	}
	wg.Wait()
	fmt.Printf("Read %d bytes in total, elapsed time: %v\n", nbytes.Load(), time.Since(start))
}

func MmapRead(filename string) {
	fp, err := os.Open(filename)
	if err != nil {
		util.Fatal(err)
	}
	defer fp.Close()

	stat, _ := fp.Stat()

	start := time.Now()

	data, err := unix.Mmap(int(fp.Fd()), 0, int(stat.Size()), unix.PROT_READ, unix.MAP_SHARED)
	if err != nil {
		util.Fatal(err)
	}
	defer unix.Munmap(data)

	fmt.Printf("read %d bytes via mmap, elapsed time: %v\n", len(data), time.Since(start))
}

func BufioRead(filename string) {
	fp, err := os.Open(filename)
	if err != nil {
		util.Fatal(err)
	}
	defer fp.Close()

	start := time.Now()

	nbytes := 0
	for r := bufio.NewReaderSize(fp, READER_BUFSIZE); ; {
		var buf [4096]byte

		n, err := r.Read(buf[:])
		if err != nil {
			if !errors.Is(err, io.EOF) {
				util.Fatal("bufio.Reader.Read:", err)
			}
			break
		}
		nbytes += n
	}

	fmt.Printf("Read %d bytes in total, elapsed time: %v\n", nbytes, time.Since(start))
}

func Readv(filename string) {
	fp, err := os.Open(filename)
	if err != nil {
		util.Fatal(err)
	}
	defer fp.Close()

	const NR_IOV = 10
	var iovs = make([][]byte, NR_IOV)
	for i := range iovs {
		iovs[i] = make([]byte, READER_BUFSIZE)
	}
	stat, _ := fp.Stat()
	start := time.Now()
	nbytes := 0
	for nbytes < int(stat.Size()) {
		n, err := unix.Readv(int(fp.Fd()), iovs)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				util.Fatal("unix.Readv: ", err)
			}
			break
		}
		nbytes += n
	}
	fmt.Printf("Read %d bytes in total, elapsed time: %v\n", nbytes, time.Since(start))
}

func ReadFile(filename string) {
	switch _options.ioengine {
	case "readv":
		Readv(filename)
	case "pread":
		PreadInParallel(filename)
	case "mmap":
		MmapRead(filename)
	case "go":
		fallthrough
	default:
		BufioRead(filename)
	}
}

func main() {
	parseCmdOptions()

	ReadFile(_options.filename)
}
