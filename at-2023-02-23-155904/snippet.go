// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-02-23 15:59:04

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"fmt"
	"time"
)

//go:generate stringer -type Region -linecomment
type Region int

const (
	RegionChina Region = iota + 3001 // china
	RegionAsia                       // asia
	RegionSea                        // sea
	RegionSA                         // sa
	RegionAmeu                       // ameu
	RegionOc                         // oc
)

type TrafficPoint struct {
	Timestamp time.Time
	Value     int64
}

type CdnProvider string

const (
	CdnBaidu      CdnProvider = "baidu"
	CdnBaishanyun CdnProvider = "baishan"
	CdnAliyun     CdnProvider = "aliyun"
)

type TrafficStat struct {
	Domain     string
	Day        time.Time
	Region     Region
	Cdn        CdnProvider
	Timeseries []TrafficPoint
}

type reduceKey struct {
	Domain string
	Day    time.Time
	Region Region
}

func AddArray(s1, s2 []TrafficPoint) []TrafficPoint {
	if len(s1) != len(s2) {
		panic("BUG: len(a) != len(b)")
	}

	var points []TrafficPoint
	var perTimeValue = map[time.Time]int64{}

	for _, v := range s1 {
		perTimeValue[v.Timestamp] += v.Value
	}
	for _, v := range s2 {
		if _, exist := perTimeValue[v.Timestamp]; !exist {
			panic("BUG: found element in s2 which is missing from s1")
		}

		perTimeValue[v.Timestamp] += v.Value
	}
	for ts, v := range perTimeValue {
		points = append(points, TrafficPoint{Timestamp: ts, Value: v})
	}
	return points
}

func aggregate(traffics []TrafficStat) []TrafficStat {
	var rtraffics []TrafficStat
	var group = map[reduceKey][]TrafficPoint{}

	for i := 0; i < len(traffics); i++ {
		traffic := traffics[i]
		key := reduceKey{Domain: traffic.Domain, Day: traffic.Day, Region: traffic.Region}

		v, exist := group[key]
		if exist {
			v = AddArray(v, traffic.Timeseries)
		} else {
			v = traffic.Timeseries
		}
		group[key] = v
	}
	for k, v := range group {
		rtraffics = append(rtraffics, TrafficStat{
			Domain:     k.Domain,
			Day:        k.Day,
			Region:     k.Region,
			Timeseries: v,
		})
	}
	return rtraffics
}

func main() {
	today := time.Now()
	traffics := []TrafficStat{
		{
			Domain: "www.example.com", Day: today, Region: RegionChina, Cdn: CdnBaidu,
			Timeseries: []TrafficPoint{
				{Timestamp: today.Add(-5 * time.Minute), Value: 1},
				{Timestamp: today.Add(-10 * time.Minute), Value: 2},
				{Timestamp: today.Add(-15 * time.Minute), Value: 3},
			},
		},
		{
			Domain: "www.example.com", Day: today, Region: RegionChina, Cdn: CdnAliyun,
			Timeseries: []TrafficPoint{
				{Timestamp: today.Add(-5 * time.Minute), Value: 3},
				{Timestamp: today.Add(-10 * time.Minute), Value: 4},
				{Timestamp: today.Add(-15 * time.Minute), Value: 5},
			},
		},
		{
			Domain: "www.example.com", Day: today, Region: RegionChina, Cdn: CdnBaishanyun,
			Timeseries: []TrafficPoint{
				{Timestamp: today.Add(-5 * time.Minute), Value: 4},
				{Timestamp: today.Add(-10 * time.Minute), Value: 5},
				{Timestamp: today.Add(-15 * time.Minute), Value: 6},
			},
		},
		{
			Domain: "www.a.example.com", Day: today, Region: RegionChina, Cdn: CdnAliyun,
			Timeseries: []TrafficPoint{
				{Timestamp: today.Add(-5 * time.Minute), Value: 3},
				{Timestamp: today.Add(-10 * time.Minute), Value: 4},
				{Timestamp: today.Add(-15 * time.Minute), Value: 5},
			},
		},
		{
			Domain: "www.a.example.com", Day: today, Region: RegionChina, Cdn: CdnBaishanyun,
			Timeseries: []TrafficPoint{
				{Timestamp: today.Add(-5 * time.Minute), Value: 4},
				{Timestamp: today.Add(-10 * time.Minute), Value: 5},
				{Timestamp: today.Add(-15 * time.Minute), Value: 6},
			},
		},
	}
	fmt.Printf("agg: %+v\n", aggregate(traffics))

	// rsets.map(x -> strings.ToUpper(x)).groupBy(x -> (x._1,x._2,x._3)).reduce(x1, x2 -> {}).collect()
}
