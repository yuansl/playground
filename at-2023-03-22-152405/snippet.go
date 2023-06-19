// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-03-22 15:24:05

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"errors"
	"fmt"
	"os"
	"time"
)

const TimeFmtMongo3 = "2006-01-02T15:04:05.999-0700"

var ErrLineTooShortToParseTime = errors.New("line too short to parse time")

func parseLocal(layout, dest string) (time.Time, error) {
	t, err := time.Parse(layout, dest)
	if err != nil {
		return t, err
	}
	t = t.Add(-time.Hour * 8)
	t = t.In(time.Local)
	return t, err
}

func parseTimeMongo3(line []byte) (t time.Time, err error) {
	if len(line) < len(TimeFmtMongo3) {
		err = ErrLineTooShortToParseTime
		return
	}
	return parseLocal(TimeFmtMongo3, string(line[0:len(TimeFmtMongo3)]))
}

func fatal(v ...any) {
	fmt.Fprintln(os.Stderr, "fatal error: ", v)
	os.Exit(1)
}

func main() {
	t, err := parseTimeMongo3([]byte(`2023-03-22T15:22:06.066+0800 I NETWORK  [initandlisten] connection accepted from 127.0.0.1:47428 #1537777 (193 connections now open)`))
	if err != nil {
		fatal(err)
	}
	fmt.Printf("t = %v, time.Local=%v\n", t, time.Now())
}
