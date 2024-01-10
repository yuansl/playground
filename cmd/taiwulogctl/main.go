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
	"regexp"
	"runtime"
	"time"

	"github.com/qbox/net-deftones/clients/sinkv2"
	"github.com/qbox/net-deftones/logger"
	netutil "github.com/qbox/net-deftones/util"
	"golang.org/x/sync/errgroup"

	titan "github.com/yuansl/playground/clients/titannetwork"
	"github.com/yuansl/playground/cmd/taiwulogctl/sinker/robotsinker"
	"github.com/yuansl/playground/cmd/taiwulogctl/taiwu"
	"github.com/yuansl/playground/cmd/taiwulogctl/taiwu/titannetwork"
	"github.com/yuansl/playground/util"
)

var _filenameTimestampRegex = regexp.MustCompile(`[[:digit:]]{12}`)

const (
	_ACCESS_KEY_DEFAULT  = "557TpseUM8ovpfUhaw8gfa2DQ0104ZScM-BTIcBx"
	_SECRET_KEY_DEFAULT  = "d9xLPyreEG59pR01sRQcFywhm4huL-XEpHHcVa90"
	_KODO_BUCKET_DEFAULT = "fusionlogtest"
)

var (
	_accessKey     = _ACCESS_KEY_DEFAULT
	_secretKey     = _SECRET_KEY_DEFAULT
	_bucket        = _KODO_BUCKET_DEFAULT
	_prefix        string
	_limit         int
	_begin         time.Time
	_end           time.Time
	_domain        string
	_outputdir     string
	_mode          string
	_robotsinkAddr string
	_version       int
	_concurrency   int
	_sink          bool
)

func init() {
	if env := os.Getenv("ACCESS_KEY"); env != "" {
		_accessKey = env
	}
	if env := os.Getenv("SECRET_KEY"); env != "" {
		_secretKey = env
	}
}

func parseCmdArgs() {
	flag.TextVar(&_begin, "begin", time.Time{}, "begin time")
	flag.TextVar(&_end, "end", time.Time{}, "end time")
	flag.StringVar(&_domain, "domain", "audiosdk.xmcdn.com", "specify cdn domain")
	flag.StringVar(&_outputdir, "dir", "./", "directory for saving logs")
	flag.StringVar(&_mode, "mode", "stat", "specify mode(one of stat|download)")
	flag.StringVar(&_bucket, "bucket", _KODO_BUCKET_DEFAULT, "specify bucket name")
	flag.StringVar(&_prefix, "prefix", "", "specify a bucket key prefix")
	flag.StringVar(&_robotsinkAddr, "robotsink", "http://xs321:30060", "address of robotsink service")
	flag.IntVar(&_version, "version", 1, "api version")
	flag.IntVar(&_concurrency, "c", runtime.NumCPU(), "concurrency")
	flag.BoolVar(&_sink, "sink", false, "sink result if set")
	flag.Parse()
}

func downloadlogs(ctx context.Context, domain string, begin, end time.Time, outputdir string, taiwu taiwu.LogService) error {
	egroup, ctx := errgroup.WithContext(ctx)
	defer egroup.Wait()

	for datetime := begin; datetime.Before(end); datetime = datetime.Add(5 * time.Minute) {
		_datetime := datetime
		links, err := taiwu.LogLinks(ctx, domain, datetime, "386BD183")
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

			egroup.Go(func() error {
				filename := path.Join(outputdir, fmt.Sprintf("/%s_%s-%04d.json", domain, _datetime.Format("200601021504"), _i))
				fp, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY|os.O_TRUNC, 0644)
				if err != nil {
					return fmt.Errorf("os.OpenFile: %w", err)
				}
				defer fp.Close()

				return netutil.WithRetry(ctx, func() error {
					if err := download(ctx, _link.Url, fp); err != nil {
						switch {
						case errors.Is(err, titan.ErrInvalid):
							panic("BUG: you shoud review your request becuase of the error(INVALID): " + err.Error())
						default:
							var cause flate.CorruptInputError
							if errors.As(err, &cause) {
								logger.FromContext(ctx).Warnf("%w: download %q error: %v, skip ...\n", cause, _link.Url, err)
								return nil
							}
							return err
						}
					}
					logger.FromContext(ctx).Infof("saved %q as %q \n", _link.Url, filename)
					return nil
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
	parseCmdArgs()

	if _begin.IsZero() {
		util.Fatal("you must specify 'begin'")
	}
	if _end.IsZero() {
		util.Fatal("you must specify 'end'")
	}

	go func() {
		runtime.SetBlockProfileRate(1)
		http.ListenAndServe(":6060", nil)
	}()

	ctx := logger.NewContext(context.TODO(), logger.New())
	switch _mode {
	case "stat":
		var trafficSrv TrafficSinker

		if _sink {
			client, err := sinkv2.NewClient(_robotsinkAddr)
			if err != nil {
				util.Fatal("sinkv2.NewClient:", err)
			}
			trafficSrv = robotsinker.NewTrafficSinker(client)
		} else {
			trafficSrv = NewNopTrafficSinker()
		}

		err := stat(ctx, flag.Args(), ProcessWindow{_begin, _end}, trafficSrv)
		if err != nil {
			util.Fatal("stat error:", err)
		}
	case "download":
		titancli := titan.NewClient(
			titan.WithCredential("qiniu", []byte("a5c90e5370c80067a2ac78aab1badb90")),
		)
		taiwusrv := titannetwork.NewTaiwuService(titancli, titannetwork.WithVersion(_version))

		err := downloadlogs(ctx, _domain, _begin, _end, _outputdir, taiwusrv)
		if err != nil {
			util.Fatal("downloadlogs error:", err)
		}
	default:
		usage()
	}
}
