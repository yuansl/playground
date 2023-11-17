// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-11-10 11:28:19

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
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/yuansl/playground/util"
)

const _NR_FIELDS_LINE = 46

type gzreader struct {
	*gzip.Reader
	underlying io.ReadCloser
}

var _ io.ReadCloser = (*gzreader)(nil)

// Close implements io.ReadCloser.
func (r *gzreader) Close() error {
	if err := r.underlying.Close(); err != nil {
		return err
	}
	return r.Reader.Close()
}

func download(ctx context.Context, url string) (io.ReadCloser, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	gz, err := gzip.NewReader(res.Body)

	return &gzreader{Reader: gz, underlying: res.Body}, err
}

func main() {
	reader, err := download(context.TODO(), `https://fusionlog.qiniu.com/v2/jdvod9xhnfjp5.vod.126.net_2023-11-02-18_part-00000.gz?e=1699585319&token=V5BwWT7pVm1S_EVHt2bfg4qOS-1VDLXCo1k6MqN1:ixjyFY1ZUJiQRrmXFbjx6cU8uHM=`)
	if err != nil {
		util.Fatal(err)
	}
	defer reader.Close()

	for lineno, r := 0, bufio.NewReader(reader); ; lineno++ {
		line, err := r.ReadBytes('\n')
		if err != nil {
			if !errors.Is(err, io.EOF) {
				util.Fatal(err)
			}
			break
		}
		line = bytes.TrimSpace(line)

		fields := bytes.Split(line, []byte{'\t'})

		if len(fields) != _NR_FIELDS_LINE {
			util.Fatal(err)
		}

		if lineno%1000 == 0 { // sample
			fmt.Printf("%d: len(fields)=%d\n", lineno, len(fields))
		}
	}
}
