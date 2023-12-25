package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/yuansl/playground/util"
)

const _TRAFFIC_SERVICE_ENDPOINT = "http://deftonestraffic.fusion.internal.qiniu.io"

type RetryUploader interface {
	Retry(ctx context.Context, domains []string, cdn string, start, end time.Time) error
}

func beginingOfday(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

type trafficService struct{}

func (*trafficService) fetchDomainTrafficOnce(ctx context.Context, domain, cdn, metric string, begin time.Time, end time.Time) ([]TrafficStat, error) {
	var result struct{ CDNLOG struct{ Points []int64 } }
	var body bytes.Buffer
	endDate := end
	if end.Sub(begin) < 24*time.Hour && end.Day() == begin.Day() {
		endDate = end.AddDate(0, 0, +1)
	}
	if cdn == "" {
		cdn = "total"
	}
	json.NewEncoder(&body).Encode(map[string]any{
		"cdn":      cdn,
		"domain":   domain,
		"start":    begin.Local().Format(time.DateOnly),
		"end":      endDate.Local().Format(time.DateOnly),
		"g":        "5min",
		"type":     metric,
		"protocol": []string{"http", "https"},
		"region":   []string{"china", "foreign"},
	})
	res, err := http.Post(_TRAFFIC_SERVICE_ENDPOINT+"/v2/admin/traffic/domain/compare", "application/json", &body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if err = json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}
	var timeseries []TrafficPoint
	begin0 := beginingOfday(begin)
	for i, it := range result.CDNLOG.Points {
		timestamp := begin0.Add(time.Duration(i) * 5 * time.Minute)

		if timestamp.Compare(begin) >= 0 && timestamp.Before(end) {
			timeseries = append(timeseries, TrafficPoint{Time: timestamp, Value: it})
		}
	}
	return []TrafficStat{{Timeseries: timeseries, Domain: domain}}, nil
}

// Traffic implements CdnTrafficStatistics.
func (srv *trafficService) Traffic(ctx context.Context, domains []string, cdn, metric string, begin time.Time, end time.Time) ([]TrafficStat, error) {
	var stats []TrafficStat

	for _, domain := range domains {
		stats0, err := srv.fetchDomainTrafficOnce(ctx, domain, cdn, metric, begin, end)
		if err != nil {
			return nil, err
		}
		stats = append(stats, stats0...)
	}
	return stats, nil
}

type logrepairer struct {
	realtimeRepairer   RetryUploader
	historicalRepairer RetryUploader
}

func (srv *logrepairer) Repair(ctx context.Context, domain, cdn string, timestamp time.Time) error {
	var repairer RetryUploader

	if timestamp.Sub(time.Now()).Abs() < 48*time.Hour {
		repairer = srv.realtimeRepairer
	} else {
		repairer = srv.historicalRepairer
	}
	return repairer.Retry(ctx, []string{domain}, cdn, timestamp, timestamp.Add(1*time.Hour))
}

func NewLogRepair(realtime, historical RetryUploader) LogRepairer {
	return &logrepairer{realtime, historical}
}

var _ CdnTrafficSerivce = (*trafficService)(nil)

type CdnTrafficServiceOption util.Option

type cdnTrafficServiceOptions struct {
	etlservice RetryUploader
}

func WithLogetlService(etlservice RetryUploader) CdnTrafficServiceOption {
	return util.OptionFunc(func(opt any) {
		opt.(*cdnTrafficServiceOptions).etlservice = etlservice
	})
}

func NewCdnTrafficService(opts ...CdnTrafficServiceOption) CdnTrafficSerivce {
	var options cdnTrafficServiceOptions
	for _, opt := range opts {
		opt.Apply(&options)
	}
	return &trafficService{}
}
