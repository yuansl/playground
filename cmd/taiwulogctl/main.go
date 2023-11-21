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
	"context"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"regexp"
	"runtime"
	"time"

	"github.com/qbox/net-deftones/clients/sinkv2"

	titan "github.com/yuansl/playground/clients/titannetwork"
	"github.com/yuansl/playground/cmd/taiwulogctl/sinker/robotsinker"
	"github.com/yuansl/playground/cmd/taiwulogctl/taiwu/titannetwork"
	"github.com/yuansl/playground/logger"
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

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: %s -mode <download|stat> -begin <RFC3339 time> `, os.Args[0])
	os.Exit(0)
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
		if err := aggregate(ctx, flag.Args(), ProcessWindow{_begin, _end}, trafficSrv); err != nil {
			util.Fatal("aggregate error:", err)
		}
	case "download":
		titancli := titan.NewClient(
			titan.WithCredential("qiniu", []byte("a5c90e5370c80067a2ac78aab1badb90")),
			titan.WithToken("386BD183"),
			titan.WithVersion(_version),
		)
		downloadlogs(ctx, _domain, _begin, _end, _outputdir, titannetwork.NewTaiwuService(titancli))
	default:
		usage()
	}
}
