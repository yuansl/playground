// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2024-01-15 17:07:47

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
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/yuansl/playground/util"
	"golang.org/x/sync/errgroup"
)

const BUFSIZE = 16 << 10

func foreachline(r io.Reader, do func(line string) error) {
	for reader := bufio.NewReaderSize(r, BUFSIZE); ; {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if !errors.Is(err, io.EOF) {
				util.Fatal("bufio.Reader.Read:", err)
			}
			break
		}
		line = bytes.TrimSpace(line)

		do(string(line))
	}
}

func run(domains []string, begin, end time.Time, w io.Writer) error {
	cmd := exec.Command("./logchecker",
		"-domains", strings.Join(domains, ","),
		"-metric", "flow",
		"-begin", begin.Format(time.RFC3339),
		"-end", end.Format(time.RFC3339),
		"-repair",
		"-uploaders", "jjh2503:7892,jjh2504:7892,jjh2792:7892,jjh3173:7892,jjh3174:7892",
		"-delta", "1",
		"-v",
		"-endpoint", "http://xs210:12324")
	cmd.Stderr = w
	cmd.Stdout = w
	return cmd.Run()
}

var _options struct {
	filename   string
	begin, end time.Time
}

func parseCmdOptions() {
	flag.TextVar(&_options.begin, "begin", time.Time{}, "begin time (in RFC3339)")
	flag.TextVar(&_options.end, "end", time.Time{}, "end time(in RFC3339)")
	flag.StringVar(&_options.filename, "f", "", "filename")
	flag.Parse()
}

func main() {
	parseCmdOptions()
	if _options.filename == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s -f <filename> -begin <...> -end <...>\n", os.Args[0])
		os.Exit(1)
	}

	fp, err := os.Open(_options.filename)
	if err != nil {
		util.Fatal(err)
	}
	defer fp.Close()

	logf, err := os.OpenFile("diff.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		util.Fatal(err)
	}
	defer logf.Close()

	var domains []string

	const BATCH_SIZE = 20

	var wg errgroup.Group

	wg.SetLimit(runtime.NumCPU())

	foreachline(fp, func(line string) error {
		if line == "domain" {
			return nil // ignore
		}
		domains = append(domains, line)
		if len(domains) >= BATCH_SIZE {
			_domains := slices.Clone(domains)
			wg.Go(func() error {
				return run(_domains, _options.begin, _options.end, logf)
			})
			domains = domains[:0]
		}
		return nil
	})
	wg.Wait()
}
