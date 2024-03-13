// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2024-03-12 15:33:08

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/qbox/net-deftones/stream"
	"github.com/yuansl/playground/util"
)

const BUFSIZE = 32 << 10

//go:generate stringer -type Metric -linecomment
type Metric int

const (
	Bandwidth Metric = iota // BANDWIDTH
	Reqcount                // REQCOUNT
)

type TrafficPoint struct {
	Time      time.Time
	Value     int64 `json:",string"`
	Countries []struct {
		Country string
		Measure struct {
			Bps int64 `json:",string"`
		}
	} `json:"-"`
}

func main() {
	var stat struct {
		Metrics []struct {
			Metric     string
			Domain     string
			Region     string
			Timeseries []TrafficPoint
		} `json:"result"`
	}
	fp, err := os.Open("/home/yuansl/Downloads/jjh2443.json")
	if err != nil {
		util.Fatal(err)
	}
	defer fp.Close()

	if err = json.NewDecoder(bufio.NewReaderSize(fp, BUFSIZE)).Decode(&stat); err != nil {
		util.Fatal(err)
	}

	for _, it := range stat.Metrics {
		if it.Metric == Reqcount.String() {
			continue
		}
		fmt.Printf("domain: %s, metric: %s, len(timeseries): %d\n", it.Domain, it.Metric, len(it.Timeseries))

		sort.Slice(it.Timeseries, func(i, j int) bool { return it.Timeseries[i].Time.Before(it.Timeseries[j].Time) })

		for i := range it.Timeseries {
			fmt.Printf("\t%+v\n", it.Timeseries[i])
		}

		stats := stream.StreamOf[TrafficPoint, time.Time, TrafficPoint](it.Timeseries).
			GroupBy(func(x TrafficPoint) time.Time { return x.Time.Local() }).
			Collect()
		for _, it2 := range stats {
			points := it2.Value.([]TrafficPoint)
			if len(points) > 1 {
				fmt.Printf("len(points)=%d\n", len(points))
				for _, p := range points {
					sum := int64(0)
					for _, it3 := range p.Countries {
						sum += it3.Measure.Bps
					}
					fmt.Printf("countries.bps.sum=%d, total=%+v\n", sum, p.Value)
				}
			}
		}
	}
}
