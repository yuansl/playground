// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-06-26 23:11:00

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

type RegionPair map[string]string

var ContinentAsRegions = map[string]string{
	"America": "AM",
	"Europe":  "em",
	"Africa":  "AF",
}

func _ContinentAsRegions() RegionPair {
	return map[string]string{
		"America": "AM",
		"Europe":  "em",
		"Africa":  "AF",
	}
}

func main() {
}
