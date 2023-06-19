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
	"strconv"
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

type Zone = string

const (
	_ZONE_CHINA   Zone = "3001"
	_ZONE_FOREIGN Zone = "3007"
)

type Request struct {
	Item    string
	Domains []string
	Uid     uint32
	Start   time.Time
	End     time.Time
	Zone    Zone
}

func send(_ context.Context, req *Request, res any) error {
	params := url.Values{}

	params.Add("select", "bandwidth")
	params.Add("start", req.Start.Format("20060102150405"))
	params.Add("end", req.End.Format("20060102150405"))
	params.Add("uid", strconv.Itoa(int(req.Uid)))
	params.Add("zone", req.Zone)
	for _, domain := range req.Domains {
		params.Add("domains", domain)
	}
	resp, err := http.Get(_FUSION_TRAFFIC_ENDPOINT + "/v2/item/" + req.Item + "?" + params.Encode())
	if err != nil {
		fatal(err)
	}

	return json.NewDecoder(resp.Body).Decode(res)
}

type Point struct {
	Time   time.Time `json:"time"`
	Values struct {
		V int64 `json:"v"`
	} `json:"values"`
}

const _X_UID = 1382514080 // "1380293517"

func GetCdnTrafficStat(ctx context.Context, domains []string, begin, end time.Time) ([]Point, error) {
	var points []Point

	for _, zone := range []string{_ZONE_CHINA, _ZONE_FOREIGN} {
		var points0 []Point

		send(ctx, &Request{
			Item:    "fusion:transfer:all:ov",
			Uid:     _X_UID,
			Domains: domains,
			Start:   begin,
			End:     end,
			Zone:    zone,
		}, &points0)
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

	send(ctx, &Request{
		Item:    "fusion:transfer:302traffic:all",
		Uid:     _X_UID,
		Domains: domains,
		Start:   begin,
		End:     end,
		Zone:    _ZONE_CHINA,
	}, &points)

	return points, nil
}

func merge(a, b []Point) []Point {
	if len(a) != len(b) {
		panic("BUG: mismatch len(a) != len(b)")
	}

	var c []Point
	var groupBy = make(map[time.Time]int64)

	sort.Slice(a, func(i, j int) bool {
		return a[i].Time.Before(a[j].Time)
	})
	sort.Slice(b, func(i, j int) bool {
		return b[i].Time.Before(b[j].Time)
	})

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

func _302trafficStat(ctx context.Context, begin, end time.Time) {
	_302peak95s := []int64{}
	_cdnpeak95s := []int64{}

	for t := begin; t.Before(end); t = t.Add(24 * time.Hour) {
		_begin, _end := t, t.Add(24*time.Hour)
		_302traffic, err := Get302CdnTrafficStat(ctx, _302domains[:], _begin, _end)
		if err != nil {
			fatal("Get302CdnTrafficStat:", err)
		}
		_cdntraffic, err := GetCdnTrafficStat(ctx, _302domains[:], _begin, _end)
		if err != nil {
			fatal("GetCdnTrafficStat:", err)
		}
		combined := merge(_302traffic, _cdntraffic)

		_302peak95s = append(_302peak95s, peak95of(combined).Values.V)

		c, err := GetCdnTrafficStat(ctx, _cdndomains, _begin, _end)
		if err != nil {
			fatal("GetCdnTrafficStat:", err)
		}
		_cdnpeak95s = append(_cdnpeak95s, peak95of(c).Values.V)
	}
	fmt.Println("_302peak95s: ", _302peak95s, "\n_cdnpeak95s:", _cdnpeak95s)
	fmt.Printf("302+兜底: %.2f Mbps(1000)\n", float64(sum(_302peak95s))/float64(int64(end.Sub(begin).Hours())/24)/1000_000)
	fmt.Printf("cdn: %.2f Mbps(1000)\n", float64(sum(_cdnpeak95s))/float64(int64(end.Sub(begin).Hours())/24)/1000_000)
}

const (
	Kibps = 1 << 10
	Mibps = 1024 * Kibps
	Gibps = 1024 * Mibps
)

func main() {
	begin := time.Date(2023, 4, 1, 0, 0, 0, 0, time.Local)
	end := begin.AddDate(0, +1, 0)

	cdntraffics, err := GetCdnTrafficStat(context.TODO(), []string{"muy.cdn.edgemec.com"}, begin, end)
	if err != nil {
		fatal("GetCdnTrafficStat:", err)
	}
	_302traffics, err := Get302CdnTrafficStat(context.TODO(), []string{"muy.cdn.edgemec.com"}, begin, end)

	c := merge(cdntraffics, _302traffics)

	peak95 := peak95of(c)
	for i, p := range c {
		fmt.Printf("%3d %s: %.2f Mibps\n", i+1, p.Time.Format("2006/01/02 15:04"), float64(p.Values.V)/Mibps)
	}
	fmt.Printf("peak95: %+v\n", peak95)
}
