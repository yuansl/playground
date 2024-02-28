// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2024-02-26 22:22:52

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	_ "embed"
	"os"
	"os/exec"

	"github.com/yuansl/playground/util"
)

//go:embed ls
var lsbin []byte

func copyExecutable(filename string) {
	fp, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		util.Fatal("os.OpenFile:", err)
	}
	defer fp.Close()
	fp.Write(lsbin)
}

func main() {
	executable := "/tmp/ls"

	copyExecutable(executable)

	defer func() {
		err := os.Remove(executable)
		if err != nil {
			util.Fatal("os.Remove:", err)
		}
	}()
	cmd := exec.Command(executable, "/")
	cmd.Stdout = os.Stdout
	cmd.Run()
	// output, err := cmd.Output()
	// if err != nil {
	// 	util.Fatal("cmd.Output:", err)
	// }
	// fmt.Println("Output:", string(output))
}
