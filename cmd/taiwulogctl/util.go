package main

import (
	"bufio"
	"compress/flate"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	netutil "github.com/qbox/net-deftones/util"
	"golang.org/x/sync/errgroup"

	"github.com/yuansl/playground/clients/titannetwork"
	"github.com/yuansl/playground/cmd/taiwulogctl/taiwu"
	"github.com/yuansl/playground/logger"
	"github.com/yuansl/playground/oss"
	"github.com/yuansl/playground/oss/kodo"
	"github.com/yuansl/playground/util"
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
		return fmt.Errorf("%w: http.Get(url=%q): %v", titannetwork.ErrProtocol, url, err)
	}
	defer res.Body.Close()

	gz, err := gzip.NewReader(res.Body)
	if err != nil {
		return fmt.Errorf("%w: gzip.NewReader(res.Body): %v", titannetwork.ErrProtocol, err)
	}
	_, err = io.Copy(saveas, gz)
	return err
}

func downloadlogs(ctx context.Context, domain string, begin, end time.Time, outputdir string, taiwu taiwu.TaiwuService) {
	egroup, ctx := errgroup.WithContext(ctx)

	for datetime := begin; datetime.Before(end); datetime = datetime.Add(5 * time.Minute) {
		_datetime := datetime
		links, err := taiwu.LogLink(ctx, domain, datetime)
		if err != nil {
			switch {
			case errors.Is(err, titannetwork.ErrInvalid):
				logger.FromContext(ctx).Warnf("taiwu.LogLink error: %v\n", err)
				return
			default:
				util.Fatal("loglink(domain=%s,datetime=%v): %v\n", domain, datetime, err)
			}
		}
		for i, link := range links {
			_i := i
			_link := link

			egroup.Go(func() error {
				log := logger.FromContext(ctx)

				log.Infof("downloading %q ...\n", _link)

				filename := path.Join(outputdir, fmt.Sprintf("/%s_%s-%04d.json", domain, _datetime.Format("200601021504"), _i))
				fp, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY|os.O_TRUNC, 0644)
				if err != nil {
					util.Fatal(err)
				}
				defer fp.Close()

				if err := netutil.WithRetry(ctx, func() error {
					if err := download(ctx, _link.Url, fp); err != nil {
						switch {
						case errors.Is(err, titannetwork.ErrInvalid):
							panic("BUG: you shoud review your request becuase of the error: " + err.Error())
						default:
							var cause flate.CorruptInputError
							if errors.As(err, &cause) {
								logger.FromContext(ctx).Warnf("download %q error: %v, skip ...\n", _link.Url, err)
								return nil
							}
							return err
						}
					}
					return nil
				}); err != nil {
					util.Fatal("download error:", err)
				}

				log.Infof("saved %q as %q \n", _link.Url, filename)
				return nil
			})
		}
	}
	egroup.Wait()
}

type pattern struct {
	domain    string
	timestamp time.Time
}

func saveAs(ctx context.Context, logs []TaiwuStandardLog, timestamp time.Time, uniq map[pattern]int, oss oss.ObjectStorageService) error {
	outfile, err := os.CreateTemp("/tmp/", "pdntaiwu_"+_domain+"*")
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
	key := pattern{domain: _domain, timestamp: timestamp}
	id, exists := uniq[key]
	if exists {
		id++
	}
	uniq[key] = id

	fkey := fmt.Sprintf("pdntaiwu/%s/%s/part-%04d.gz", timestamp.Format("2006-01-02-15"), _domain, id)

	outfile.Seek(0, io.SeekStart)

	_, err = oss.Upload(ctx, "defy-etl-log", outfile, kodo.WithKey(fkey))
	return err
}
