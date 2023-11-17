package main

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/qbox/net-deftones/clients/sinkv2"
	"github.com/qbox/net-deftones/fusionrobot"
	"github.com/qbox/net-deftones/staging/qiniu.com/cdnapi/fusion.v2/fusion"

	"github.com/yuansl/playground/logger"
)

type TrafficSinker interface {
	Sink(ctx context.Context, _ []TrafficStat) error
}

type TrafficStat struct {
	Timeseries []TrafficPoint
	Day        time.Time
	Domain     string
	Region     Region
	DataType   DataType
}

type robotsinker struct {
	*sinkv2.Client
}

const _NR_POINTS_PER_DAY = 288

// Sink implements TrafficService.
func (sinker *robotsinker) Sink(ctx context.Context, stats []TrafficStat) error {
	var perDayTraffics = make(map[groupKey][]TrafficStat)

	for _, stat := range stats {
		key := groupKey{Domain: stat.Domain, Day: stat.Day}
		perDayTraffics[key] = append(perDayTraffics[key], stat)
	}
	for key, stats := range perDayTraffics {
		var points []fusionrobot.RawDayTrafficTidb

		for i := 0; i < len(stats); i++ {
			stat := &stats[i]

			sort.Slice(stat.Timeseries, func(i, j int) bool {
				return stat.Timeseries[i].Timestamp.Before(stat.Timeseries[j].Timestamp)
			})

			bandwidths := make([]int64, _NR_POINTS_PER_DAY)

			for j := 0; j < len(stat.Timeseries); j++ {
				ts := &stat.Timeseries[j]

				logger.FromContext(ctx).Infof("%s: bytes=%d\n", ts.Timestamp, ts.Bytes)

				index := ts.Timestamp.Hour()*12 + ts.Timestamp.Minute()/5
				if index >= len(bandwidths) {
					panic(fmt.Sprintf("BUG: index(%d) > _NR_POINTS_PER_DAY(%d)", index, _NR_POINTS_PER_DAY))
				}
				bandwidths[index] = ts.Bytes * 8 / 300
			}
			points = append(points, fusionrobot.RawDayTrafficTidb{
				Domain:     key.Domain,
				Day:        key.Day.Format(time.DateOnly),
				Region:     fusion.GeoCoverChina,
				SourceType: fusionrobot.SourceType("LOG"),
				DataType:   fusionrobot.DataTypePDNBandwidth,
				CDN:        fusion.CDNProvider("pdntaiwu"),
				Data:       bandwidths,
			})
		}

		traffic := sinkv2.SavePointsRequest{RequestId: "yuansl-" + key.Domain + key.Day.Format(time.DateOnly), DayPoints: points}

		logger.FromContext(ctx).Infof("saving traffic: %+v\n", &traffic)

		if err := sinker.SaveDayTraffic(ctx, &traffic); err != nil {
			return err
		}
	}
	return nil
}

var _ TrafficSinker = (*robotsinker)(nil)

func NewTrafficSinker(client *sinkv2.Client) TrafficSinker {
	return &robotsinker{Client: client}
}
