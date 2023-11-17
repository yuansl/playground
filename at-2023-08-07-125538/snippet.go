// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-08-07 12:55:38

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strconv"
)

func fatal(v ...any) {
	pc, file, line, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)

	fmt.Fprintln(os.Stderr, "fatal error: ", v)
	fmt.Fprintf(os.Stderr, "%s:\n\t%s:%d\n", fn.Name(), file, line)
	os.Exit(1)
}

func main() {
	var v map[string]any

	x := strconv.Quote("abc\n")
	qx, err := strconv.Unquote(x)
	if err != nil {
		fatal("strconv.Unquote:", err)
	}
	qqx, _ := strconv.Unquote(`"` + x[1:len(x)-1] + `"`)
	fmt.Printf("qx='%s', x='%d', qqx='%s'\n", qx, len(x), qqx)

	// const raw = `{"from-edge-https":"test35","uicode":"test37"}`

	const raw = "\"{\\x22from-edge-https\\x22:\\x22test35\\x22,\\x22uicode\\x22:\\x22test37\\x22,\\x22idTag\\x22:\\x22sina\\x22,\\x22cookie\\x22:\\x22test12\\x22,\\x22edge-rid\\x22:\\x22test26\\x22,\\x22x-cdn-from\\x22:\\x22qn\\x22,\\x22x-log-uid\\x22:\\x22test11\\x22,\\x22x-log-biztype\\x22:\\x22test42\\x22,\\x22x-log-oid\\x22:\\x22test28\\x22,\\x22luicode\\x22:\\x22test39\\x22,\\x22x-log-videotype\\x22:\\x22test29\\x22,\\x22lfid\\x22:\\x22test40\\x22,\\x22x-log-sessionid\\x22:\\x22test30\\x22,\\x22fid\\x22:\\x22test38\\x22}\""

	unquoted, err := strconv.Unquote(raw)
	if err != nil {
		fatal("strconv.Unquote:", err)
	}

	fmt.Printf("unquoted: '%s'\n", unquoted)

	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		fatal("json.Unmarshal:", err)
	}

	fmt.Println("v:", v, unquoted)

}
