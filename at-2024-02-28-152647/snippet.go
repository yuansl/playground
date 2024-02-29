// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2024-02-28 15:26:47

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
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/yuansl/playground/util"
)

type TrafficPoint struct {
	Time       time.Time
	Value      int64 `json:",string"`
	Dimensions []struct {
		Country string
		Measure struct {
			Bps int64 `json:",string"`
		}
	}
}

type TrafficStat struct {
	Metric     string
	Domain     string
	Region     string
	Timeseries []TrafficPoint
}

type TrafficStatResponse struct {
	Result []TrafficStat
}

const (
	KiB = 1024
	MiB = 1024 * KiB
	GiB = 1024 * MiB
	TiB = 1024 * GiB
)

var _options struct {
	filename string
}

func parseCmdOptions() {
	flag.StringVar(&_options.filename, "file", "", "specify traffic result file in json formt")
	flag.Parse()
}

func main() {
	parseCmdOptions()
	var result TrafficStatResponse

	fp, err := os.Open(_options.filename)
	if err != nil {
		util.Fatal("os.Open:", err)
	}
	defer fp.Close()

	err = json.NewDecoder(bufio.NewReader(fp)).Decode(&result)
	if err != nil {
		util.Fatal("json.Decode:", err)
	}
	var traffic int64
	var chinatraffic = make(map[time.Time]int64)
	for _, r := range result.Result {
		if r.Metric == "BANDWIDTH" {
			for _, ts := range r.Timeseries {
				traffic += ts.Value
				for _, dim := range ts.Dimensions {
					if dim.Country == "CN" {
						chinatraffic[ts.Time.UTC()] += dim.Measure.Bps
					}
				}
			}
		}
	}

	fmt.Printf("total traffic at 20240225: %d Bytes, traffic in china mainland:\n", traffic*300/8)
	for t, bps := range chinatraffic {
		fmt.Printf("%s: %d\n", t, bps*300/8)
	}
}
