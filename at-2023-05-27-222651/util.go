package main

import (
	"bufio"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"os"
)

func fatal(v ...any) {
	fmt.Fprintln(os.Stderr, "fatal error: ", v)
	os.Exit(1)
}

func md5sumOf(filename string) []byte {
	fp, err := os.Open(filename)
	if err != nil {
		fatal("os.Open:", err)
	}
	defer fp.Close()
	r := bufio.NewReader(fp)
	md5sum := md5.New()
	for {
		buf := [_BLOCK_SIZE]byte{}

		n, err := r.Read(buf[:])
		if err != nil {
			if !errors.Is(err, io.EOF) {
				fatal("bufio.Read:", err)
			}
			break
		}
		md5sum.Write(buf[:n])
	}
	sum := md5sum.Sum(nil)
	return sum[:]
}
