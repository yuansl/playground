// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-07-24 19:25:33

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
	"regexp"

	cmp "golang.org/x/exp/constraints"
)

func Scale[S ~[]E, E cmp.Integer](s S, c E) S {
	for i := 0; i < len(s); i++ {
		s[i] *= c
	}
	return s
}

type Serializer[T any] interface {
	Serialize(T) ([]byte, error)
}

type Point []int32

func (p Point) String() string { return fmt.Sprintf("Point{%d,%d,%d}", p[0], p[1], p[2]) }

func (p Point) Serialize(Point) ([]byte, error) {
	return []byte(fmt.Sprintf("Point{%d,%d,%d}", p[0], p[1], p[2])), nil
}

func foo2[T Serializer[T]](x T) {

}

func IndexFunc[S ~[]E, E any](s S, f func(E) bool) int {
	for i, v := range s {
		if f(v) {
			return (i)
		}
	}
	return -1
}

type T struct {
	_ any
	_ [0]func()
}

type T2 struct {
	_ any
	_ [0]*func()
}

func isComparable[_ comparable]() {

}

func fatal(v ...any) {
	fmt.Fprintln(os.Stderr, "fatal error: ", v)
	os.Exit(1)
}

func main() {
	const raw = `{\\x22from-edge-https\\x22:\\x22test35\\x22,\\x22uicode\\x22:\\x22test37\\x22,\\x22x-cdn-from\\x22:\\x22qn\\x22,\\x22cookie\\x22:\\x22test12\\x22,\\x22x-log-biztype\\x22:\\x22test42\\x22,\\x22edge-rid\\x22:\\x22test26\\x22,\\x22luicode\\x22:\\x22test39\\x22,\\x22x-log-uid\\x22:\\x22test11\\x22,\\x22lfid\\x22:\\x22test40\\x22,\\x22x-log-oid\\x22:\\x22test28\\x22,\\x22fid\\x22:\\x22test38\\x22,\\x22x-log-videotype\\x22:\\x22test29\\x22,\\x22idTag\\x22:\\x22sina\\x22,\\x22x-log-sessionid\\x22:\\x22test30\\x22}`

	// s := strings.NewReplacer(`\\x22`, `"`).
	// 	Replace(raw)

	regex := regexp.MustCompile("\\\\\\\\x22")

	s0 := regex.ReplaceAll([]byte(raw), []byte(`"`))

	fmt.Printf("s0 = '%s'\n", s0)

	var v map[string]any

	if err := json.Unmarshal(s0, &v); err != nil {
		fatal("json.Unmarshal error:", err)
	}

	fmt.Printf("v=%v\n", v)
}
