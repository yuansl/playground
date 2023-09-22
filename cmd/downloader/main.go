package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const _RANGE_SIZE = 4 << 20 // 4MiB

var (
	ErrInvalid  = errors.New("invalid argument")
	ErrProtocol = errors.New("protocol error")
)

type GetOption Option
type getOptions struct {
	rangeBegin int
	rangeEnd   int
}

func WithRange(begin, end int) GetOption {
	return OptionFn(func(op any) {
		op.(*getOptions).rangeBegin = begin
		op.(*getOptions).rangeEnd = end
	})
}

func Get(ctx context.Context, url string, opts ...GetOption) ([]byte, error) {
	var options getOptions

	for _, op := range opts {
		op.apply(&options)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: http.NewRequest: %v", ErrInvalid, err)
	}
	if options.rangeBegin >= 0 && options.rangeEnd > options.rangeBegin {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", options.rangeBegin, options.rangeEnd-1))
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http.Client.Do: %v", err)
	}
	defer res.Body.Close()

	return io.ReadAll(res.Body)
}

type DownloadOption Option

type downloadOptions struct {
	saveAs      string
	concurrency int
}

func WithSaveAs(outfile string) DownloadOption {
	return OptionFn(func(op any) {
		op.(*downloadOptions).saveAs = outfile
	})
}

func WithConcurrency(concurrency int) DownloadOption {
	return OptionFn(func(op any) {
		op.(*downloadOptions).concurrency = concurrency
	})
}

type DiscardFileWriter struct{}

func (*DiscardFileWriter) Sync() error {
	return nil
}

func (*DiscardFileWriter) Name() string {
	return ""
}

func (*DiscardFileWriter) WriteAt(p []byte, off int64) (n int, err error) {
	return len(p), nil
}

var _ WriteSyncer = (*DiscardFileWriter)(nil)

type WriteSyncer interface {
	io.WriterAt
	Name() string
	Sync() error
}

func Download(ctx context.Context, url string, opts ...DownloadOption) error {
	var options downloadOptions
	var w WriteSyncer

	for _, op := range opts {
		op.apply(&options)
	}
	if options.concurrency <= 0 {
		options.concurrency = runtime.NumCPU()
	}
	if options.saveAs != "" {
		tmpf, err := os.CreateTemp("/tmp", filepath.Base(options.saveAs)+"-*.tmp")
		if err != nil {
			return fmt.Errorf("os.CreateTemp: %v", err)
		}
		defer tmpf.Close()
		w = tmpf
	} else {
		w = &DiscardFileWriter{}
	}
	var downbytes atomic.Int64

	rangeRequest := func(rbegin, rend int) error {
		rawbytes, err := Get(ctx, url, WithRange(rbegin, rend))
		if err != nil {
			return fmt.Errorf("%w: Get(): %v", ErrProtocol, err)
		}
		if _, err := w.WriteAt(rawbytes, int64(rbegin)); err != nil {
			return fmt.Errorf("w.WriteAt: %v", err)
		}
		downbytes.Add(int64(rend - rbegin))
		return nil
	}
	res, err := http.Head(url)
	if err != nil {
		return fmt.Errorf("%w: http.HEAD: %v", ErrProtocol, err)
	}
	defer res.Body.Close()

	size, _ := strconv.Atoi(res.Header.Get("Content-Length"))

	if strings.Contains(res.Header.Get("Accept-Ranges"), "bytes") {
		var wg sync.WaitGroup
		var climit = make(chan struct{}, options.concurrency)

		go func() {
			for range time.Tick(1 * time.Second) {
				fmt.Printf("progress: %d/%d(%.2f%%)\n", downbytes.Load(), size, float64(downbytes.Load())/float64(size)*100)
			}
		}()
		for off := 0; off < size; off += _RANGE_SIZE {
			climit <- struct{}{}
			wg.Add(1)
			go func(off int) {
				defer func() {
					<-climit
					wg.Done()
				}()

				rbegin, rend := off, off+_RANGE_SIZE
				if rend > size {
					rend = size
				}
				rangeRequest(rbegin, rend)
			}(off)
		}
		wg.Wait()
	} else {
		rangeRequest(0, -1)
	}

	if options.saveAs != "" {
		if err = w.Sync(); err != nil {
			return fmt.Errorf("os.File.Sync: %v", err)
		}
		if err := os.Rename(w.Name(), options.saveAs); err != nil {
			return fmt.Errorf("os.Rename: %v", err)
		}
	}
	return nil
}

var (
	ouptut      = flag.String("o", "/dev/null", "specify output file name")
	concurrency = flag.Int("c", runtime.NumCPU(), "concurrency")
)

func main() {
	flag.Parse()

	if len(os.Args) < 2 {
		fatal("Usage: %s <url> [-o output]\n", os.Args[0])
	}

	start := time.Now()
	defer func() { fmt.Printf("time elapsed %v\n", time.Since(start)) }()

	if err := Download(context.TODO(), os.Args[len(os.Args)-1], WithSaveAs(*ouptut), WithConcurrency(*concurrency)); err != nil {
		fatal("Download error: ", err)
	}
}
