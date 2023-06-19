// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-06-12 17:26:58

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"sync"
	"time"
)

var _302domains = [...]string{"bs-c.resource.ccplay.cn", "bs.ccplay-resource.zhuifenghanhua.cn", "p1.resource.ccplay.cn", "p4.resource.ccplay.cn", "apk-open1.ccplay.cn", "p3.resource.ccplay.cn", "bs4.resource.ccplay.cn"}

var _alldomains = [...]string{
	"i3.resource.ccplay.cc",
	"i4.resource.ccplay.cc",
	"resource.ccplay.cc",
	"i1.resource.ccplay.cc",
	"bs.ccplay-resource.zhuifenghanhua.cn",
	"i4.resource.ccplay.cn",
	"i2.resource.ccplay.cc",
	"oversea.ccplay.mobi",
	"resource.androidswiki.net",
	"resource.playmods.net",
	"ccplay-resource.zhuifenghanhua.cn",
	"p4.resource.ccplay.com.cn",
	"static-resource.ccplay.cn",
	"i4-resource.ccplay.cn",
	"p1.resource.ccplay.cn",
	"sitemaps.playmods.net",
	"i3.resource.ccplay.cn",
	"ccplay-resource.ccplaydis.com",
	"i1.resource.ccplay.cn",
	"p4.resource.ccplay.cn",
	"i1-resource.ccplay.cn",
	"i2.resource.ccplay.cn",
	"p-resource.ccplay.cn",
	"resource.ccplay.cn",
	"archive.heibaige.com",
	"resource.rowtechapk.com",
	"i2-resource.ccplay.cn",
	"resource.popularmodapk.com",
	"station.playmods.net",
	"static.ccplay.cn",
	"cp.resource.zhuifenghanhua.cn",
	"resource.latestmodsapk.com",
	"i3-resource.ccplay.cn",
	"p2.resource.ccplay.cc",
	"apk-open.playmods.net",
	"p2.resource.ccplay.cn",
	"resource.ccplay.com",
	"p3.resource.ccplay.cn",
	"qn-cdn.zhuifenghanhua.com",
	"cp.resource.ccplaydis.com",
	"resource.playmods.app",
	"apk-p-resource.ccplay.cn",
	"ccp.qn-res.zhuifenghanhua.cn",
	"resource.androidswiki.com",
	"resource.ccplay.com.cn",
	"resource.funmod.online",
	"d3.media.ccplay.com.cn",
	".ccplay.com.cn",
	"bs4.resource.ccplay.cn",
	"static.ccplay.cc",
	"i3.media.ccplay.com.cn",
	"p-c.resource.ccplay.cn",
	"resource.playmods.top",
	"open-resource.ccplay.cn",
	"ios-resource.ccplay.cc",
	"bs-c.resource.ccplay.cn",
	"tmp.resource.ccplay.cn",
	"apk-open1.ccplay.cn",
	"ccp.qn-res.ccplaydis.com",
	"upres.media.ccplaydis.com",
	"qn.ccplay-resource.ccplaydis.com",
	"videoclip-resource.playmods.net",
	"p4-resource.ccplay.cn",
	"p1-resource.ccplay.cn",
	"p3-resource.ccplay.cn",
	"apk-i1-resource.ccplay.cn",
	"apk-p3-resource.ccplay.cn",
	"apk-p4-resource.ccplay.cn",
	"apk-p1-resource.ccplay.cn",
	"apk-i2-resource.ccplay.cn",
	"apk-p2-resource.ccplay.cn",
	"apk-i3-resource.ccplay.cn",
	"p2-resource.ccplay.cn",
	"apk-i4-resource.ccplay.cn",
}

var _cdndomains []string

var _initOnce sync.Once

func init() {
	_initOnce.Do(func() {
		var _302domainsdict = make(map[string]struct{})

		for _, d := range _302domains {
			_302domainsdict[d] = struct{}{}
		}
		for _, domain := range _alldomains {
			if _, exists := _302domainsdict[domain]; !exists {
				_cdndomains = append(_cdndomains, domain)
			}
		}
	})
}

const _FUSION_TRAFFIC_ENDPOINT = "http://xs394:20051"

func fatal(v ...any) {
	fmt.Fprintln(os.Stderr, "fatal error: ", v)
	os.Exit(1)
}

type Request struct {
	Domains []string
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Response struct {
	Result any `json:"result"`
	*Error `json:",omitempty"`
}

func sendRequest(item string, domains []string, begin, end time.Time, zone string, v any) error {
	keys := url.Values{}

	keys.Add("uid", "1380293517")
	keys.Add("zone", zone)
	keys.Add("select", "bandwidth")
	keys.Add("start", begin.Format("20060102150405"))
	keys.Add("end", end.Format("20060102150405"))
	for _, domain := range domains {
		keys.Add("domains", domain)
	}
	resp, err := http.Get(_FUSION_TRAFFIC_ENDPOINT + "/v2/item/" + item + "?" + keys.Encode())
	if err != nil {
		fatal(err)
	}

	return json.NewDecoder(resp.Body).Decode(v)
}

type Point struct {
	Time   time.Time `json:"time"`
	Values struct {
		V int64 `json:"v"`
	} `json:"values"`
}

func GetCdnTrafficStat(ctx context.Context, domains []string, begin, end time.Time) ([]Point, error) {
	var points []Point

	for _, zone := range []string{"3001", "3007"} {
		var points0 []Point

		sendRequest("fusion:transfer:all:ov", domains, begin, end, zone, &points0)
		if len(points) == 0 {
			points = points0
		} else {
			points = merge(points, points0)
		}
	}

	return points, nil
}

func Get302CdnTrafficStat(ctx context.Context, domains []string, begin, end time.Time) ([]Point, error) {
	var points []Point

	sendRequest("fusion:transfer:302traffic:all", domains, begin, end, "3001", &points)

	return points, nil
}

func merge(a, b []Point) []Point {
	if len(a) != len(b) {
		panic("BUG: mismatch len(a) != len(b)")
	}

	var c []Point
	var groupBy = make(map[time.Time]int64)

	for i, v := range a {
		groupBy[v.Time] += v.Values.V + b[i].Values.V
	}
	for t, v := range groupBy {
		c = append(c, Point{Time: t, Values: struct {
			V int64 `json:"v"`
		}{V: v}})
	}
	return c
}

func peak95of(points []Point) Point {
	sort.Slice(points, func(i, j int) bool {
		return points[i].Values.V > points[j].Values.V
	})
	peak95index := len(points) / 20
	if peak95index > 0 {
		peak95index--
	}
	if peak95index < 0 {
		return Point{Values: struct {
			V int64 `json:"v"`
		}{V: -1}}
	}
	return points[peak95index]
}

func sum(nums []int64) int64 {
	var s int64
	for _, num := range nums {
		s += num
	}
	return s
}

func main() {
	_302peak95s := []int64{}
	_cdnpeak95s := []int64{}
	begin := time.Date(2023, 5, 1, 0, 0, 0, 0, time.Local)
	end := time.Date(2023, 6, 1, 0, 0, 0, 0, time.Local)
	ctx := context.TODO()

	for t := begin; t.Before(end); t = t.Add(24 * time.Hour) {
		_302traffic, err := Get302CdnTrafficStat(ctx, _302domains[:], t, t.Add(24*time.Hour))
		if err != nil {
			fatal("Get302CdnTrafficStat:", err)
		}
		_cdntraffic, err := GetCdnTrafficStat(ctx, _302domains[:], t, t.Add(24*time.Hour))
		if err != nil {
			fatal("GetCdnTrafficStat:", err)
		}
		combined := merge(_302traffic, _cdntraffic)
		_302peak95s = append(_302peak95s, peak95of(combined).Values.V)

		c, err := GetCdnTrafficStat(ctx, _cdndomains, t, t.Add(24*time.Hour))
		if err != nil {
			fatal("GetCdnTrafficStat:", err)
		}
		_cdnpeak95s = append(_cdnpeak95s, peak95of(c).Values.V)
	}
	fmt.Println("_302peak95s: ", _302peak95s, "\n_cdnpeak95s:", _cdnpeak95s)
	fmt.Printf("302+兜底: %.2f Mbps(1000)\n", float64(sum(_302peak95s))/float64(int64(end.Sub(begin).Hours())/24)/1000_000)
	fmt.Printf("cdn: %.2f Mbps(1000)\n", float64(sum(_cdnpeak95s))/float64(int64(end.Sub(begin).Hours())/24)/1000_000)
}
