// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-12-26 18:55:34

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"context"
	"fmt"
	"sort"
	"time"

	fusionutil "github.com/qbox/net-deftones/fusionrobot/util"

	"github.com/yuansl/playground/util"
)

type TrafficPoint struct {
	Timestamp time.Time
	Value     int64
}

type TrafficService interface {
	GetCdnTraffic(ctx context.Context, cdn string, begin, end time.Time) ([]TrafficPoint, error)
	GetDomainsTraffic(ctx context.Context, cdn string, domains []string, begin, end time.Time) ([]TrafficPoint, error)
}

func dissection(set1, set2 []TrafficPoint) []TrafficPoint {
	perTimePoint := func(set []TrafficPoint) map[time.Time]int64 {
		var m = make(map[time.Time]int64)
		for _, it := range set {
			m[it.Timestamp.UTC()] = it.Value
		}
		return m
	}
	m1 := perTimePoint(set1)
	m2 := perTimePoint(set2)
	for t := range m1 {
		m1[t] -= m2[t]
	}

	var dist = make([]TrafficPoint, 0, len(m1))
	for t, v := range m1 {
		dist = append(dist, TrafficPoint{Timestamp: t, Value: v})
	}
	return dist
}

func ValuesOf(points []TrafficPoint) []int64 {
	var values = make([]int64, 0, len(points))

	for _, p := range points {
		values = append(values, p.Value)
	}
	return values
}

func assert(expression bool, msg ...string) {
	if !expression {
		panic("BUG: " + fmt.Sprint(msg))
	}
}

const NR_POINTS_PER_DAY = 288

func peak95of(points []TrafficPoint) TrafficPoint {

	assert(len(points) == NR_POINTS_PER_DAY, fmt.Sprintf("len(points) must be %d, but got %d", NR_POINTS_PER_DAY, len(points)))

	sort.Slice(points, func(i, j int) bool {
		return points[i].Value > points[j].Value
	})
	peak95index := len(points) / 20
	if peak95index > 0 {
		peak95index -= 1
	}

	return points[peak95index]
}

func sum(points []TrafficPoint) int64 {
	return fusionutil.Sum(ValuesOf(points))
}

var options struct {
	cdn     string
	uid     uint
	domains []string
	begin   time.Time
	end     time.Time
}

func stat(peak95s, dispeak95s []TrafficPoint, begin, end time.Time) float64 {
	cost := float64(sum(peak95s)) / (end.Sub(begin).Hours() / 24)
	cost0 := float64(sum(dispeak95s)) / (end.Sub(begin).Hours() / 24)
	delta := cost - cost0

	fmt.Printf("(avrPeak95) cost: %.4f , cost(without uid %d): %.4f, delta: %.4f\n", cost, options.uid, cost0, delta)

	return delta
}

func main() {
	var service TrafficService
	var ctx = context.TODO()
	var peak95s []TrafficPoint
	var dispeak95s []TrafficPoint

	for date := options.begin; date.Before(options.end); date = date.AddDate(0, 0, +1) {
		points1, err := service.GetCdnTraffic(ctx, options.cdn, date, date.AddDate(0, 0, +1))
		if err != nil {
			util.Fatal("service.GetCdnTraffic: ", err)
		}
		peak95s = append(peak95s, peak95of(points1))

		points2, err := service.GetDomainsTraffic(ctx, options.cdn, options.domains, options.begin, options.end)
		if err != nil {
			util.Fatal("service.GetDomainsTraffic: ", err)
		}
		dis := dissection(points1, points2)
		dispeak95s = append(dispeak95s, peak95of(dis))
	}

	stat(peak95s, dispeak95s, options.begin, options.end)
}
