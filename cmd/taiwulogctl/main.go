// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-11-13 10:35:05

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"compress/flate"
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/qbox/net-deftones/clients/sinkv2"
	"github.com/qbox/net-deftones/logger"
	"github.com/qbox/net-deftones/util"
	netutil "github.com/qbox/net-deftones/util"
	"golang.org/x/sync/errgroup"

	titan "github.com/yuansl/playground/clients/titannetwork"
	pcdn "github.com/yuansl/playground/cmd/taiwulogctl/logservice"
	"github.com/yuansl/playground/cmd/taiwulogctl/logservice/titannetwork"
	"github.com/yuansl/playground/cmd/taiwulogctl/sinker/robotsinker"
)

const (
	_ACCESS_KEY_DEFAULT  = "557TpseUM8ovpfUhaw8gfa2DQ0104ZScM-BTIcBx"
	_SECRET_KEY_DEFAULT  = "d9xLPyreEG59pR01sRQcFywhm4huL-XEpHHcVa90"
	_KODO_BUCKET_DEFAULT = "fusionlogtest"
)

var _options struct {
	accessKey     string
	secretKey     string
	bucket        string
	prefix        string
	limit         int
	begin         time.Time
	end           time.Time
	domain        string
	outputdir     string
	mode          string
	robotsinkAddr string
	version       int
	concurrency   int
	sink          bool
	timeout       time.Duration
}

func init() {
	_options.accessKey = _ACCESS_KEY_DEFAULT
	_options.secretKey = _SECRET_KEY_DEFAULT

	if env := os.Getenv("ACCESS_KEY"); env != "" {
		_options.accessKey = env
	}
	if env := os.Getenv("SECRET_KEY"); env != "" {
		_options.secretKey = env
	}
}

func parseCmdOptions() {
	flag.TextVar(&_options.begin, "begin", time.Time{}, "begin time")
	flag.TextVar(&_options.end, "end", time.Time{}, "end time")
	flag.StringVar(&_options.domain, "domain", "audiosdk.xmcdn.com", "specify cdn domain")
	flag.StringVar(&_options.outputdir, "dir", "./", "directory for saving logs")
	flag.StringVar(&_options.mode, "mode", "stat", "specify mode(one of stat|download)")
	flag.StringVar(&_options.bucket, "bucket", _KODO_BUCKET_DEFAULT, "specify bucket name")
	flag.StringVar(&_options.prefix, "prefix", "", "specify a bucket key prefix")
	flag.StringVar(&_options.robotsinkAddr, "robotsink", "http://xs321:30060", "address of robotsink service")
	flag.IntVar(&_options.version, "version", 1, "api version")
	flag.IntVar(&_options.concurrency, "c", runtime.NumCPU(), "concurrency")
	flag.BoolVar(&_options.sink, "sink", false, "sink result if set")
	flag.DurationVar(&_options.timeout, "timeout", 10*time.Minute, "specify timeout")
	flag.Parse()
}

const _XIAOHONGSHU_TAIWU_TOKEN = "386BD183"

func downloadlogs(ctx context.Context, domain string, begin, end time.Time, outputdir string, logservice pcdn.LogService) error {
	wg, ctx := errgroup.WithContext(ctx)
	defer wg.Wait()

	wg.SetLimit(_options.concurrency)

	for datetime := begin; datetime.Before(end); datetime = datetime.Add(5 * time.Minute) {
		_datetime := datetime
		links, err := logservice.LogLinks(ctx, domain, datetime, _XIAOHONGSHU_TAIWU_TOKEN)
		if err != nil {
			switch {
			case errors.Is(err, titan.ErrInvalid):
				return fmt.Errorf("taiwu.LogLink: %w", err)
			default:
				return fmt.Errorf("loglink(domain=%s,datetime=%v): %v", domain, datetime, err)
			}
		}
		for i, link := range links {
			_i := i
			_link := link

			wg.Go(func() error {
				tmpf, err := os.CreateTemp("./", "*.tmp")
				if err != nil {
					return fmt.Errorf("os.OpenFile: %w", err)
				}
				defer tmpf.Close()

				log := logger.FromContext(ctx)

				return netutil.WithRetry(ctx, func() error {
					err := download(ctx, _link.Url, tmpf)
					if err != nil {
						switch {
						case errors.Is(err, titan.ErrInvalid):
							log.Warn("download(url=%d) error: %v, skipped\n", err)
							return nil

						default:
							var cause flate.CorruptInputError
							if errors.As(err, &cause) {
								log.Warnf("%w: download %q error: %v, skip ...\n", cause, _link.Url, err)
								return nil
							}
							return err
						}
					}

					filename := path.Join(outputdir, fmt.Sprintf("/%s_%s-%04d.json", domain, _datetime.Format("200601021504"), _i))
					log.Infof("saved %q as %q \n", _link.Url, filename)

					return os.Rename(tmpf.Name(), filename)
				})
			})
		}
	}
	return nil
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: %s -mode <download|stat> -begin <RFC3339 time> `, os.Args[0])
	os.Exit(1)
}

func main() {
	parseCmdOptions()

	if _options.begin.IsZero() {
		util.Fatal("you must specify 'begin'")
	}
	if _options.end.IsZero() {
		util.Fatal("you must specify 'end'")
	}

	go func() {
		runtime.SetBlockProfileRate(1)
		http.ListenAndServe(":6060", nil)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), _options.timeout)
	defer cancel()

	ctx = logger.NewContext(ctx, logger.New())

	switch _options.mode {
	case "stat":
		var trafficSrv TrafficSinker

		if _options.sink {
			client, err := sinkv2.NewClient(_options.robotsinkAddr)
			if err != nil {
				util.Fatal("sinkv2.NewClient:", err)
			}
			trafficSrv = robotsinker.NewTrafficSinker(client)
		} else {
			trafficSrv = NewNopTrafficSinker()
		}

		err := stat(ctx, flag.Args(), ProcessWindow{_options.begin, _options.end}, trafficSrv)
		if err != nil {
			util.Fatal("stat error:", err)
		}
	case "download":
		titancli := titan.NewClient(
			titan.WithCredential("qiniu", []byte("a5c90e5370c80067a2ac78aab1badb90")),
		)
		taiwusrv := titannetwork.NewTaiwuLogService(titancli, titannetwork.WithVersion(_options.version))

		err := downloadlogs(ctx, _options.domain, _options.begin, _options.end, _options.outputdir, taiwusrv)
		if err != nil {
			util.Fatal("downloadlogs error:", err)
		}
	default:
		usage()
	}
}
