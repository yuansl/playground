// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-12-12 18:37:56

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
	"math"
	"os"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	fusionutil "github.com/qbox/net-deftones/util"
	"github.com/yuansl/playground/util"
)

type UploadStat struct {
	Domain string
	Cdn    string
	Start  time.Time
	End    time.Time
	Bytes  int64
}

type CdnlogUploader interface {
	Stat(ctx context.Context, domains []string, cdn string, begin, end time.Time) ([]UploadStat, error)
}

type TrafficPoint struct {
	Time  time.Time
	Value int64
}

type TrafficStat struct {
	Domain     string
	Timeseries []TrafficPoint
}

type CdnTrafficSerivce interface {
	Traffic(ctx context.Context, domains []string, cdn, metric string, begin, end time.Time) ([]TrafficStat, error)
}

type LogRepairer interface {
	Repair(ctx context.Context, domain, cdn string, timestamp time.Time) error
}

type Stat struct {
	Domain       string
	Time         time.Time
	UploadBytes  int64
	TrafficBytes int64
}

var _options struct {
	begin       time.Time
	end         time.Time
	domains     string
	cdn         string
	metric      string
	repair      bool
	verbose     bool
	etlendpoint string
	delta       float64
}

const _DELTA_WATERMARK = 1e-4

func parseCmdArgs() {
	flag.TextVar(&_options.begin, "begin", time.Time{}, "begin time(in RFC3339)")
	flag.TextVar(&_options.end, "end", time.Time{}, "end time (in RFC3339)")
	flag.StringVar(&_options.domains, "domains", "", "specify domains(seperated by comma)")
	flag.StringVar(&_options.cdn, "cdn", "dnlivestream", "cdn provider name")
	flag.StringVar(&_options.metric, "metric", "flow", "specify traffic metric(one of 302liveflow|flow)")
	flag.BoolVar(&_options.repair, "repair", false, "whether repair if there are delta between upload and traffic bytes")
	flag.BoolVar(&_options.verbose, "v", false, "verbose")
	flag.StringVar(&_options.etlendpoint, "endpoint", _LOGETL_ENDPOINT_DEFAULT, "specify logetl service endpoint")
	flag.Float64Var(&_options.delta, "delta", _DELTA_WATERMARK, "delta in percent")
	flag.Parse()
}

func stat(ctx context.Context, domains []string, trafficSrv CdnTrafficSerivce, uploaderSrv CdnlogUploader, repairer LogRepairer) {
	perTimeStat := make(map[time.Time]Stat)
	statq := make(chan *Stat)

	go func() {
		defer close(statq)

		wg, ctx := errgroup.WithContext(ctx)

		for day := _options.begin; day.Before(_options.end); day = day.AddDate(0, 0, +1) {
			begin, end := day, day.AddDate(0, 0, +1)
			if now := time.Now(); end.After(now) {
				end = now
			}
			wg.Go(func() error {
				upstats, err := uploaderSrv.Stat(ctx, domains, _options.cdn, begin, end)
				if err != nil {
					panic("BUG: uploader.Stat:" + err.Error())
				}
				for _, it := range upstats {
					statq <- &Stat{UploadBytes: it.Bytes, Domain: it.Domain, Time: it.Start}
				}

				return nil
			})

			wg.Go(func() error {
				trafficstats, err := trafficSrv.Traffic(ctx, domains, _options.cdn, _options.metric, begin, end)
				if err != nil {
					panic("BUG: trafficSrv.Traffic:" + err.Error())
				}
				for _, it := range trafficstats {
					for _, it2 := range it.Timeseries {
						statq <- &Stat{TrafficBytes: it2.Value, Time: it2.Time, Domain: it.Domain}
					}
				}
				return nil
			})
		}
		wg.Wait()
	}()
	for it := range statq {
		old := perTimeStat[it.Time.UTC()]
		old.UploadBytes += it.UploadBytes
		old.TrafficBytes += it.TrafficBytes
		perTimeStat[it.Time.UTC()] = old
	}
	var perHour = make(map[time.Time]struct{})
	for ts, stat := range perTimeStat {
		if _options.verbose {
			fmt.Printf("timestamp: %s, traffic bytes: %15d, upload bytes: %15d\n", ts, stat.TrafficBytes, stat.UploadBytes)
		}

		if stat.UploadBytes == 0 || math.Abs(float64(stat.TrafficBytes-stat.UploadBytes))/float64(stat.UploadBytes) > _options.delta {
			fmt.Fprintf(os.Stderr, "WARNING: traffic doesn't match at %s: upload %15d(bytes), traffic: %15d(bytes), delta: %.3f%%\n", ts.Local(), stat.UploadBytes, stat.TrafficBytes, math.Abs(float64(stat.TrafficBytes-stat.UploadBytes))/float64(stat.UploadBytes)*100)

			if _options.repair {
				y, m, d := ts.Date()
				perHour[time.Date(y, m, d, ts.Hour(), 0, 0, 0, ts.Location())] = struct{}{}
			}
		}
	}
	if _options.repair {
		for hour := range perHour {
			if err := fusionutil.WithRetry(ctx, func() error {
				return repairer.Repair(ctx, domains[0], _options.cdn, hour)
			}); err != nil {
				util.Fatal("repair:", err)
			}
		}
	}
}

func main() {
	parseCmdArgs()
	if _options.domains == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s -domains <domains> -begin <begin> -end <end>", os.Args[0])
		os.Exit(1)
	}
	if _options.begin.IsZero() {
		_options.begin = time.Now()
	}
	if _options.end.IsZero() {
		_options.end = time.Now()
	}
	if _options.end.After(time.Now()) {
		_options.end = time.Now()
	}
	domains := strings.Split(_options.domains, ",")
	uploaderSrv := NewCdnlogUploader()
	trafficSrv := NewCdnTrafficService()
	ctx := context.TODO()

	stat(ctx, domains, trafficSrv, uploaderSrv, NewLogRepair(uploaderSrv.(*uploaderService), NewLogetlService(WithLogetlEndpoint(_options.etlendpoint))))
}
