package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	"golang.org/x/exp/constraints"
)

func Sum[T constraints.Ordered](a ...T) T {
	var s T

	for _, v := range a {
		s += v
	}
	return s
}

func fatal(v ...any) {
	fmt.Fprintln(os.Stderr, "fatal error: ", v)
	os.Exit(1)
}

func fatalf(format string, v ...any) {
	fmt.Fprintf(os.Stderr, "fatal error: "+format, v...)
	os.Exit(1)
}

// date,uid,class,scenario
func loadUIDsFrom(filename string) []uint32 {
	uids := []uint32{}
	fp, err := os.Open(filename)
	if err != nil {
		fatal("os.Open:", err)
	}
	defer fp.Close()

	for reader := bufio.NewReader(fp); ; {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if !errors.Is(err, io.EOF) {
				fatal("bufio.ReadBytes error:", err)
			}
			break
		}
		fields := bytes.Split(line, []byte(","))
		if len(fields) < 2 {
			fatalf("invalid format of uid file: %q, len(fields)=%d\n", line, len(fields))
		}
		xuid := bytes.TrimSpace(fields[1])
		uid, err := strconv.Atoi(string(xuid))
		if err != nil {
			if string(fields[1]) == "uid" {
				continue
			}
			fatalf("invalid format of uid line: fields=%q, line=%q\n", fields[1], line)
		}
		uids = append(uids, uint32(uid))
	}
	return uids
}

func deduplicateOf[T comparable](sets []T) []T {
	newsets := make([]T, 0, len(sets))
	uniq := map[T]struct{}{}

	for _, set := range sets {
		uniq[set] = struct{}{}
	}
	for key := range uniq {
		newsets = append(newsets, key)
	}
	return newsets
}
