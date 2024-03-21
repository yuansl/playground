// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2024-03-03 11:19:53

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
	"unsafe"

	"github.com/qbox/net-deftones/util"
)

func alignoftime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
}

func main() {
	fp, err := os.Open("diff.log")
	util.Assert(fp != nil, "os.Open:", err)
	defer fp.Close()

	type DomainHour struct {
		Domain string
		Hour   time.Time
	}

	domains := util.NewSet[DomainHour]()

	for reader := bufio.NewReaderSize(fp, 32<<10); ; {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if !errors.Is(err, io.EOF) {
				util.Fatal("bufio.Reader.ByteBytes:", err)
			}
			break
		}
		if !bytes.Contains(line, []byte("WARNING")) || !bytes.Contains(line, []byte("delta: +")) {
			continue
		}
		line = bytes.TrimSpace(line)
		start := bytes.Index(line, []byte("mismatch:"))
		line = line[start+9:]
		stop := bytes.Index(line, []byte("upload"))
		line = line[:stop]
		line = bytes.ReplaceAll(line, []byte("+0800 CST:"), []byte(""))
		line = bytes.ReplaceAll(line, []byte(","), []byte(""))
		line = bytes.TrimSpace(line)
		fields := bytes.SplitN(line, []byte(" "), 3)
		domain := bytes.TrimSpace(fields[2])
		hour := append(fields[0], append([]byte{' '}, fields[1]...)...)

		timestamp, _ := time.ParseInLocation(time.DateTime, *(*string)(unsafe.Pointer(&hour)), time.Local)

		domains.Add(DomainHour{Domain: string(domain), Hour: alignoftime(timestamp)})
	}

	for _, x := range domains.List() {
		fmt.Printf("%s %s\n", x.Domain, x.Hour.Format("2006-01-02-15"))
	}
}
