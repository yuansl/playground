package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/qbox/net-deftones/logger"
)

var ErrUnknown = errors.New("mongodb: unknown error")

type Timesereis = []int64

type TrafficService interface {
	GetTimeseriesOf(ctx context.Context, cdn string, from, to time.Time, domains ...string) (Timesereis, error)
}

type cdnTrafficService struct {
	endpoint string
}

func newCdnTrafficService() TrafficService {
	return &cdnTrafficService{endpoint: "http://deftonestraffic.fusion.internal.qiniu.io"}
}

func (srv *cdnTrafficService) getTrafficOfDomain(ctx context.Context, domain string, from, to time.Time, cdn ...string) []int64 {
	req := map[string]any{
		"domain":   domain,
		"start":    from.Format(time.DateOnly),
		"end":      to.Format(time.DateOnly),
		"type":     "pdnflow",
		"g":        "5min",
		"protocol": []string{"http", "https"},
		"region":   []string{"china"},
	}
	if len(cdn) > 0 {
		req["cdn"] = cdn[0]
	}
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(req)
	if err != nil {
		fatal("json.Encode:", err)
	}

	res, err := http.Post(srv.endpoint+"/v2/admin/traffic/domain/compare", "application/json", &buf)
	if err != nil {
		fatal("http.Post:", err)
	}
	defer res.Body.Close()

	var payload struct {
		CdnLog struct {
			Points []int64
		} `json:"CDNLOG"`
	}
	if err = json.NewDecoder(res.Body).Decode(&payload); err != nil {
		fatal("json.Decode:", err)
	}

	return payload.CdnLog.Points
}

func (srv *cdnTrafficService) GetTimeseriesOf(ctx context.Context, cdn string, from, to time.Time, domains ...string) (Timesereis, error) {
	trafficq := make(chan []int64, concurrency)
	go func() {
		defer close(trafficq)
		var wg sync.WaitGroup

		for _, domain := range domains {
			wg.Add(1)
			go func(domain string) {
				defer wg.Done()
				trafficq <- srv.getTrafficOfDomain(ctx, domain, from, to, cdn)

			}(domain)
		}
		wg.Wait()
	}()

	var points Timesereis

	for _points := range trafficq {
		if len(points) == 0 {
			points = _points
			continue
		}
		points = AddArrayto(points, _points)
	}
	return points, nil
}

type vdnTrafficService struct {
	*mongo.Collection
}

func newVdnTrafficService() TrafficService {
	db, err := mongo.Connect(context.TODO(),
		options.Client().
			ApplyURI("mongodb://pili:PiLIVdN@10.30.38.29:3708,10.30.38.30:3708,10.30.38.31:3708/pili?authSource=pili"))
	if err != nil {
		fatal("mogno.Connect failed:", err)
	}
	return &vdnTrafficService{db.Database("pili").Collection("off_down_5min")}
}

func (srv *vdnTrafficService) find(ctx context.Context, idc string, from, to time.Time, domains ...string) chan []int64 {
	timeseries := make(chan []int64, 2)
	go func() {
		defer close(timeseries)

		filter := bson.M{"time": bson.M{"$gte": from, "$lt": to}, "idc": cdn}
		if len(domains) > 0 {
			filter["domain"] = bson.M{"$in": domains}
		}
		it, err := srv.Find(ctx, filter)
		if err != nil {
			fatal("%w: %w", ErrUnknown, err)
		}
		defer it.Close(ctx)

		for it.Next(ctx) {
			var doc struct {
				Flow []int64 `bson:"flow"`
			}
			if err = it.Decode(&doc); err != nil {
				fatal("%w: cursor.Decode: %v", ErrUnknown, err)
			}
			timeseries <- doc.Flow
		}
	}()
	return timeseries
}

func (srv *vdnTrafficService) GetTimeseriesOf(ctx context.Context, cdn string, from, to time.Time, domains ...string) (Timesereis, error) {
	var points Timesereis

	for timeseries := range srv.find(ctx, cdn, from, to, domains...) {
		if len(points) == 0 {
			points = timeseries
			continue
		}
		points = AddArrayto(points, timeseries)
	}

	return points, nil
}

func statPeak95(ctx context.Context, repository TrafficService, cdn string, from, to time.Time, domains ...string) int64 {
	flatpoints := []int64{}

	for day := from; day.Before(to); day = day.AddDate(0, 0, +1) {
		points, err := repository.GetTimeseriesOf(ctx, cdn, day, day.AddDate(0, 0, +1), domains...)
		if err != nil {
			fatal("repository.GetTimeserisStatOf: %v", err)
		}
		if len(points) == 0 {
			points = make([]int64, 288)
		}
		flatpoints = append(flatpoints, points...)
	}
	if len(flatpoints) == 0 {
		return 0
	}
	return int64(peak95of(flatpoints))
}

func channelOfPeaks(ctx context.Context, repository TrafficService, cdn string, from, to time.Time, domains ...string) chan int64 {
	var peaksq = make(chan int64, concurrency)

	go func() {
		defer close(peaksq)

		var wg sync.WaitGroup

		for day := from; day.Before(to); day = day.AddDate(0, 0, +1) {
			wg.Add(1)
			go func(day time.Time) {
				defer wg.Done()

				points, err := repository.GetTimeseriesOf(ctx, cdn, day, day.AddDate(0, 0, +1), domains...)
				if err != nil {
					fatal("repository.GetTimeseriesStatOf error:", err)
				}

				peaksq <- peakOf(points)
			}(day)
		}
		wg.Wait()
	}()
	return peaksq
}

func statAveragePeakOf(ctx context.Context, r TrafficService, cdn string, from, to time.Time, domains ...string) float64 {
	peaks := []int64{}
	for peak := range channelOfPeaks(ctx, r, cdn, from, to, domains...) {
		peaks = append(peaks, peak)
	}
	days := to.Sub(from).Hours() / 24

	return float64(sum(peaks)*8/300) / float64(days)
}

const Unit = 1000

const (
	Kbps = Unit
	Mbps = Unit * Kbps
	Gbps = Unit * Mbps
	Tbps = Unit * Gbps
	Pbps = Unit * Tbps
)

var (
	from        time.Time
	to          time.Time
	domain      string
	cdn         string
	concurrency int
	averagePeak bool
)

func parseCmdArgs() {
	flag.TextVar(&from, "from", time.Now(), "start time")
	flag.TextVar(&to, "to", time.Now(), "end time")
	flag.BoolVar(&averagePeak, "avrpeak", false, "if compute average peak")
	flag.StringVar(&domain, "domain", "", "specify domains")
	flag.StringVar(&cdn, "cdn", "pdntaiwu", "specify the cdn provider")
	flag.IntVar(&concurrency, "c", runtime.NumCPU(), "concurrency")
	flag.Parse()
}

func main() {
	parseCmdArgs()

	domains := []string{}
	if domain != "" {
		domains = strings.Split(domain, ",")
	}
	repo := newCdnTrafficService()
	ctx := logger.NewContext(context.TODO(), logger.New())
	peak95 := statPeak95(ctx, repo, cdn, from, to, domains...)
	fmt.Printf("peak95: %d Mbps\n", peak95*8/300/Mbps)

	if averagePeak {
		avrPeak95 := statAveragePeakOf(ctx, repo, cdn, from, to, domains...)
		fmt.Printf("average peak: %s, %.3f Mbps\n", from.Format("2006-01"), avrPeak95/Mbps)
	}
}
