// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-09-12 17:22:58

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/yuansl/playground/utils"
)

var fatal = utils.Fatal

func SaveAs(fp *os.File) {

	if data, err := os.ReadFile(fp.Name()); err != nil {
		fatal("os.ReadFile:", err)
	} else {
		fmt.Printf("Read from %s: %q\n", fp.Name(), data)
	}

	if st, err := os.Stat(fp.Name()); err != nil {
		fatal("os.Stat:", err)
	} else {
		fmt.Printf("File Size:%d, inode: %d\n", st.Size(), st.Sys().(*syscall.Stat_t).Ino)
	}

	if err := os.Rename(fp.Name(), filepath.Dir(fp.Name())+"/some.log"); err != nil {
		fatal("os.Rename:", err)
	}

	data, err := os.ReadFile("/tmp/some.log")
	if err != nil {
		fatal("os.ReadFile:", err)
	}

	fmt.Printf("Read from some.log: %q\n", data)
}

// Deprecated this is a fake version mkstemp(3)
func mkstemp() {
	fp, err := os.CreateTemp("/tmp", "some.*.tmp")
	if err != nil {
		fatal("os.CreateTemp:", err)
	}
	defer fp.Close()

	fp.WriteString("I am writing to a temp file")

	fp.Sync()

	runtime.SetFinalizer(fp, func(fp *os.File) { fmt.Printf("Removing file %[1]s ...\n", fp.Name()); os.Remove(fp.Name()) })

	fmt.Printf("fp.name=%[2]s\n", fp.Fd(), fp.Name())
}

func main() {
	mkstemp()

	runtime.GC()
}
