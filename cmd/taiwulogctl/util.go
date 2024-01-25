package main

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/yuansl/playground/clients/titannetwork"
	"github.com/yuansl/playground/oss"
	"github.com/yuansl/playground/oss/kodo"
)

const (
	KiB = 1024
	MiB = 1024 * KiB
	GiB = 1024 * MiB
	TiB = 1024 * GiB
	PiB = 1024 * TiB
)

func prettyOfBytes(bytes float64) string {
	if gib := bytes / GiB; gib >= 1 {
		return fmt.Sprintf("%.2f GiB", gib)
	}
	if mib := bytes / MiB; mib >= 1 {
		return fmt.Sprintf("%.2f MiB", mib)
	}
	if kib := bytes / KiB; kib >= 1 {
		return fmt.Sprintf("%.2f KiB", kib)
	}
	return fmt.Sprintf("%.0f B", bytes)
}

func download(ctx context.Context, url string, saveas io.Writer) error {
	res, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("%w: http.Get(url=%q): %w", titannetwork.ErrProtocol, url, err)
	}
	defer res.Body.Close()

	if res.StatusCode >= http.StatusBadRequest {
		var cause error
		switch {
		case res.StatusCode >= http.StatusInternalServerError:
			cause = titannetwork.ErrUnavailable
		case res.StatusCode >= http.StatusBadRequest:
			cause = titannetwork.ErrInvalid
		default:
			cause = titannetwork.ErrProtocol
		}
		return fmt.Errorf("%w: status: '%s'", cause, res.Status)
	}
	gz, err := gzip.NewReader(res.Body)
	if err != nil {
		return fmt.Errorf("%w: gzip.NewReader(res.Body): %v", titannetwork.ErrProtocol, err)
	}
	_, err = io.Copy(saveas, gz)
	return err
}

type pattern struct {
	domain    string
	timestamp time.Time
}

func saveAs(ctx context.Context, logs []TaiwuStandardLog, timestamp time.Time, uniq map[pattern]int, oss oss.ObjectStorageService) error {
	outfile, err := os.CreateTemp("/tmp/", "pdntaiwu_"+_options.domain+"*")
	if err != nil {
		return fmt.Errorf("os.CreateTemp: %v", err)
	}
	defer outfile.Close()

	buffered := bufio.NewWriter(outfile)

	gz := gzip.NewWriter(buffered)
	for _, it := range logs {
		if err = json.NewEncoder(gz).Encode(it); err != nil {
			return fmt.Errorf("json.Encode: %w", err)
		}
	}
	gz.Close()
	if err = buffered.Flush(); err != nil {
		return fmt.Errorf("bufio.Flush: %w", err)
	}
	timestamp = time.Date(timestamp.Year(), timestamp.Month(), timestamp.Day(), timestamp.Hour(), 0, 0, 0, timestamp.Location())
	key := pattern{domain: _options.domain, timestamp: timestamp}
	id, exists := uniq[key]
	if exists {
		id++
	}
	uniq[key] = id

	fkey := fmt.Sprintf("pdntaiwu/%s/%s/part-%04d.gz", timestamp.Format("2006-01-02-15"), _options.domain, id)

	outfile.Seek(0, io.SeekStart)

	_, err = oss.Upload(ctx, "defy-etl-log", outfile, kodo.WithKey(fkey))
	return err
}
