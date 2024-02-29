// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-11-28 18:47:53

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/qbox/net-deftones/util"
)

const _NR_STAT_PER_DAY = 288

type Stat struct {
	Timestamp time.Time
	Value     int64
}

type StatService interface {
	StatOf(ctx context.Context, domain string, begin, end time.Time) ([]Stat, error)
}

type trafficService struct{}

type TrafficRequest struct {
	Start    string   `json:"start"`
	End      string   `json:"end"`
	G        string   `json:"g"`
	Domain   string   `json:"domain"`
	Protocol []string `json:"protocol"`
	Region   []string `json:"region"`
}

// StatOf implements StatService.
func (*trafficService) StatOf(ctx context.Context, domain string, begin time.Time, end time.Time) ([]Stat, error) {
	var payload bytes.Buffer

	if err := json.NewEncoder(&payload).Encode(TrafficRequest{
		Start:    begin.Format(time.DateOnly),
		End:      end.Add(24 * time.Hour).Format(time.DateOnly),
		G:        "5min",
		Domain:   domain,
		Protocol: []string{"http", "https"},
		Region:   []string{"china", "foreign"}}); err != nil {
		//
		util.Fatal(err)
	}

	var payload2 struct {
		PerCdnStat map[string]struct {
			AvrPeak   float64
			AvrPeak95 float64
			Peak      int64
			Peak95    float64
			Points    []int64
		} `json:"cdnProviders"`
	}

	if err := util.WithRetry(ctx, func() error {
		res, err := http.Post("http://deftonestraffic.fusion.internal.qiniu.io/v2/admin/traffic/domain", "application/json", &payload)
		if err != nil {
			util.Fatal(err)
		}
		defer res.Body.Close()

		return json.NewDecoder(res.Body).Decode(&payload2)
	}); err != nil {
		util.Fatal(err)
	}

	var stats []Stat

	if v, exists := payload2.PerCdnStat["aliyun"]; exists {
		y, m, d := begin.Date()
		beginofday := time.Date(y, m, d, 0, 0, 0, 0, begin.Location())
		for i, bps := range v.Points {
			timestamp := beginofday.Add(time.Duration(i) * 5 * time.Minute)
			if timestamp.Compare(begin) >= 0 && timestamp.Before(end) {
				stats = append(stats, Stat{
					Timestamp: timestamp,
					Value:     bps,
				})
			}
		}
	}

	return stats, nil
}

var _ StatService = (*trafficService)(nil)

func peakof(stats []Stat) *Stat {
	if len(stats) == 0 {
		return nil
	}
	max := &stats[0]
	for i := 1; i < len(stats); i++ {
		stat := &stats[i]
		if stat.Value > max.Value {
			max = stat
		}
	}
	return max
}

func peak95of(stats []Stat) *Stat {
	if len(stats) == 0 {
		return nil
	}

	sort.Slice(stats, func(i, j int) bool { return stats[i].Value > stats[j].Value })

	index := len(stats) / 20
	if index > 0 {
		index--
	}
	if index >= 0 {
		return &stats[index]
	}

	return nil
}

const (
	Ki = 1000
	Mi = 1000 * Ki
	Gi = 1000 * Mi
)

var (
	_domains string
)

func parseCmdArgs() {
	flag.StringVar(&_domains, "domains", "", "specify domains")
	flag.Parse()
}

func main() {
	parseCmdArgs()

	if _domains = strings.TrimSpace(_domains); _domains == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s <domains>(seperate by comma)", os.Args[0])
		os.Exit(1)
	}

	fmt.Println("domain,datetime,bandwidth/Mbps,peak95/Mbps")

	var srv StatService = &trafficService{}

	today := time.Date(2023, 11, 28, 18, 0, 0, 0, time.Local)
	ctx := context.TODO()
	begin, end := today, today.Add(1*time.Hour)
	y, m, d := today.Add(-24 * time.Hour).Date()
	yesterday := time.Date(y, m, d, 0, 0, 0, 0, time.Local)

	domains := strings.Split(_domains, ",")

	fmt.Printf("do traffic stat for %d domains ...\n", len(domains))

	for _, domain := range domains {
		stats0, err := srv.StatOf(ctx, domain, begin, end)
		if err != nil {
			util.Fatal(err)
		}
		peak := peakof(stats0)

		if peak == nil {
			peak = &Stat{}
		}

		stats1, err := srv.StatOf(ctx, domain, yesterday, yesterday.AddDate(0, 0, +1))
		if err != nil {
			util.Fatal(err)
		}
		peak95 := peak95of(stats1)
		if peak95 == nil {
			peak95 = &Stat{}
		}
		fmt.Printf("%s,%s,%d,%d\n", domain, today.Format(time.DateOnly), peak.Value/Mi, peak95.Value/Mi)
	}
}
