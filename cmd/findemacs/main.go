// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2024-03-23 20:06:36

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"os/exec"

	"github.com/yuansl/playground/util"
)

// func ListActiveWindows() {
// 	var buf bytes.Buffer
// 	cmd := exec.Command("gdbus", "call", "--session", "--dest", "org.gnome.Shell", "--object-path", "/org/gnome/Shell/Extensions/WindowsExt", "--method", "org.gnome.Shell.Extensions.WindowsExt.List")
// 	cmd.Stdout = &buf
// 	cmd.Stderr = &buf
// 	if err := cmd.Run(); err != nil {
// 		util.Fatal(err)
// 	}
// 	buf1 := bytes.Replace(buf.Bytes(), []byte(`('`), []byte(""), 1)
// 	if i := bytes.LastIndex(buf1, []byte(`'`)); i >= 0 {
// 		buf1 = buf1[:i]
// 	}

// 	var windows []struct {
// 		Class     string
// 		PID       int64
// 		ID        int64
// 		Maximized int64
// 		Focus     bool
// 		Title     string
// 	}
// 	if err := json.Unmarshal(buf1, &windows); err != nil {
// 		util.Fatal(err)
// 	}
// 	for i := range windows {
// 		if w := &windows[i]; w.Class == "emacs" {
// 			fmt.Printf("%+v\n", w)
// 		}
// 	}
// }

func RaiseEmacsWindow() error {
	cmd := exec.Command("gdbus", "call", "--session", "--dest", "org.gnome.Shell", "--object-path", "/org/gnome/Shell/Extensions/WindowsExt", "--method", "org.gnome.Shell.Extensions.WindowsExt.RaiseEmacsWindow")
	return cmd.Run()
}

func main() {
	if err := RaiseEmacsWindow(); err != nil {
		util.Fatal(err)
	}
}
