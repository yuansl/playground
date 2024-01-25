package robotsinker

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/qbox/net-deftones/clients/sinkv2"
	"github.com/qbox/net-deftones/fusionrobot"
	"github.com/qbox/net-deftones/logger"
	"github.com/qbox/net-deftones/staging/qiniu.com/cdnapi/fusion.v2/fusion"

	"github.com/yuansl/playground/cmd/taiwulogctl/sinker"
)

const _NR_POINTS_PER_DAY = 288

type robotsinker struct {
	*sinkv2.Client
}

type TrafficStat = sinker.TrafficStat

type GroupKey struct {
	Domain    string
	Timestamp time.Time
}

// Sink implements TrafficService.
func (sinker *robotsinker) Sink(ctx context.Context, stats []TrafficStat) error {
	var perDayTraffics = make(map[GroupKey][]TrafficStat)

	for _, stat := range stats {
		key := GroupKey{Domain: stat.Domain, Timestamp: stat.Day}
		perDayTraffics[key] = append(perDayTraffics[key], stat)
	}

	log := logger.FromContext(ctx)

	shuffle := func(stat *TrafficStat) []int64 {
		var bandwidths = make([]int64, _NR_POINTS_PER_DAY)

		sort.Slice(stat.Timeseries, func(i, j int) bool {
			return stat.Timeseries[i].Timestamp.Before(stat.Timeseries[j].Timestamp)
		})
		for j := 0; j < len(stat.Timeseries); j++ {
			ts := &stat.Timeseries[j]

			log.Infof("%s: bytes=%d\n", ts.Timestamp, ts.Bytes)

			index := ts.Timestamp.Hour()*12 + ts.Timestamp.Minute()/5

			if index >= len(bandwidths) {
				panic(fmt.Sprintf("BUG: index(%d) > _NR_POINTS_PER_DAY(%d)", index, _NR_POINTS_PER_DAY))
			}
			bandwidths[index] = ts.Bytes * 8 / 300
		}
		return bandwidths
	}
	for groupBy, stats := range perDayTraffics {
		var points []sinkv2.TrafficStat

		for i := 0; i < len(stats); i++ {
			bandwidths := shuffle(&stats[i])

			points = append(points, sinkv2.TrafficStat{
				Domain:     groupBy.Domain,
				Day:        groupBy.Timestamp.Format(time.DateOnly),
				Region:     string(fusion.GeoCoverChina),
				SourceType: fusionrobot.SourceTypeLog.String(),
				DataType:   int64(fusionrobot.DataTypePDNBandwidth),
				CDN:        "pdntaiwu",
				Bandwidths: bandwidths,
			})
		}

		traffic := sinkv2.SavePointsRequest{
			RequestId: logger.IdFromContext(ctx),
			DayPoints: points,
		}
		log.Infof("saving traffic: %+v\n", &traffic)

		if err := sinker.SaveDayTraffic(ctx, &traffic); err != nil {
			return err
		}
	}
	return nil
}

var _ sinker.TrafficSinker = (*robotsinker)(nil)

func NewTrafficSinker(client *sinkv2.Client) sinker.TrafficSinker {
	return &robotsinker{Client: client}
}
