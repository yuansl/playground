package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/qbox/net-deftones/util"
)

//go:generate stringer -type MetricType -linecomment
type MetricType int

const (
	MetricBandwidth    MetricType = iota // bandwidth
	MetricSrcBandwidth                   // srcbandwidth
	MetricQps                            // qps
)

func (m MetricType) IsValid() bool {
	return MetricBandwidth <= m && m <= MetricQps
}

func MetricOf(metric string) MetricType {
	for m := MetricBandwidth; m <= MetricQps; m++ {
		if m.String() == metric {
			return m
		}
	}
	return -1
}

type RangeRequest interface {
	GetTimeRangeTraffic(ctx context.Context, domains []string, cdn string, metric MetricType, begin, end time.Time)
}

var options struct {
	domains []string
	cdn     string
	metric  string
	begin   time.Time
	end     time.Time
}

func parseCmdOptions() {
	var domain, begin, end string
	var err error

	flag.StringVar(&domain, "domains", "", "specify domains, seperate by comma")
	flag.StringVar(&options.cdn, "cdn", "", "specify cdn (e.g.: qiniudcdn)")
	flag.StringVar(&options.metric, "metric", "bandwidth", "specify metric. one of (metric, srcbandwidth, qps)")
	flag.StringVar(&begin, "begin", "", "begin time(in format ccyy-mm-dd)")
	flag.StringVar(&end, "end", "", "end time (in ccyy-mm-dd)")
	flag.Parse()

	if domain != "" {
		options.domains = strings.Split(domain, ",")
	}

	options.begin, err = time.ParseInLocation(time.DateOnly, begin, time.Local)
	if err != nil {
		util.Fatal(err)
	}
	options.end, err = time.ParseInLocation(time.DateOnly, end, time.Local)
	if err != nil {
		util.Fatal(err)
	}
}

const BATCH_SIZE_MAX = 2000

func main() {
	parseCmdOptions()

	if len(options.domains) == 0 || options.cdn == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s -domains <domains> ", os.Args[0])
		os.Exit(0)
	}

	var rangerequest RangeRequest

	for i := 0; i < len(options.domains); i += BATCH_SIZE_MAX {
		j := i + BATCH_SIZE_MAX
		if j > len(options.domains) {
			j = len(options.domains)
		}

		rangerequest.GetTimeRangeTraffic(context.TODO(), options.domains[i:j], options.cdn, MetricOf(options.metric), options.begin, options.end)
	}
}
