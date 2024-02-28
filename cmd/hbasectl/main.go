// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2024-01-31 16:43:10

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/qbox/net-deftones/fscdn.v2/types"
	"github.com/qbox/net-deftones/fusionrobot"
	"github.com/qbox/net-deftones/logger"
	"golang.org/x/sync/errgroup"
)

type Stat struct {
	Domain    string
	Cdn       types.CDNProvider
	Region    fusionrobot.Region
	Timestamp time.Time
	Value     int
}

type CdnTrafficRepository interface {
	DeleteAll(ctx context.Context, domain string, cdn types.CDNProvider, region fusionrobot.Region, begin, end time.Time) error
	Stat(ctx context.Context, domain string, cdn types.CDNProvider, region fusionrobot.Region, start, end time.Time) ([]Stat, error)
}

type CdnLogStat struct {
	Cdns      []string
	Domain    string
	Timestamp time.Time
}

type CdnLogRepository interface {
	Stat(ctx context.Context, cdn types.CDNProvider, start, end time.Time) ([]CdnLogStat, error)
}

//go:generate stringer -type OpMode -linecomment
type OpMode int

const (
	STAT_MODE   OpMode = iota // stat
	DELETE_MODE               // delete
)

func inspectCdnLogTraffic(ctx context.Context, domain string, begin, end time.Time, cdn types.CDNProvider, op OpMode, store CdnTrafficRepository) error {
	wg, ctx := errgroup.WithContext(ctx)
	log := logger.FromContext(ctx)

	for region := fusionrobot.RegionAMEU; region <= fusionrobot.RegionOC; region++ {
		wg.Go(func() error {
			switch op {
			case STAT_MODE:
				stats, err := store.Stat(ctx, domain, cdn, region, begin, end)
				if err != nil {
					return fmt.Errorf("store.Stat: %w", err)
				}
				for _, p := range stats {
					log.Infof("traffic stat: %+v\n", p)
				}

			case DELETE_MODE:
				log.Debugf("delete domain=%s,cdn=%s,region=%s,begin=%s,end=%s\n", domain, cdn, region, begin, end)
				return store.DeleteAll(ctx, domain, cdn, region, begin, end)
			}
			return nil
		})
	}
	return wg.Wait()
}

func inspectCdnLogRepository(ctx context.Context, cdns []string, begin, end time.Time, repo CdnLogRepository) error {
	wg, ctx := errgroup.WithContext(ctx)
	var count atomic.Int32

	defer func() {
		logger.FromContext(ctx).Infof("cdn_hour stat: len(stat)=%d\n", count.Load())
	}()

	for _, cdn := range cdns {
		wg.Go(func() error {
			stat, err := repo.Stat(ctx, types.CDNProvider(cdn), begin, end)
			if err != nil {
				return fmt.Errorf("LogRepository.Stat: %+v\n", err)
			}
			count.Add(int32(len(stat)))

			return nil
		})
	}
	return wg.Wait()
}

var _options struct {
	domain string
	begin  time.Time
	end    time.Time
	cdn    string
	host   string
	mode   OpMode
}

func parseOptions() {
	flag.StringVar(&_options.domain, "domains", "file.yalla.live", "specify domains(sperate by comma)")
	flag.TextVar(&_options.begin, "begin", time.Time{}, "specify begin time in RFC3339 format")
	flag.TextVar(&_options.end, "end", time.Time{}, "specify end time in RFC3339 format")
	flag.StringVar(&_options.cdn, "cdn", string(types.CDNProviderBaishanyun), "specify cdn provider(seperate by comma)")
	flag.StringVar(&_options.host, "host", "localhost", "specify hbase address")
	mode := flag.String("mode", STAT_MODE.String(), "op mode")
	flag.Parse()

	if mode != nil {
		switch *mode {
		case "", STAT_MODE.String():
			_options.mode = STAT_MODE
		case DELETE_MODE.String():
			_options.mode = DELETE_MODE
		}
	}
}

func main() {
	parseOptions()

	hbase := OpenHbase(_options.host)
	logfilerepo := NewCdnLogRepository(hbase)
	// trafficrepo := NewCdnLogTraffic(hbase)
	ctx := logger.NewContext(context.Background(), logger.New())

	defer hbase.Close()

	inspectCdnLogRepository(ctx, strings.Split(_options.cdn, ","), _options.begin, _options.end, logfilerepo)

	// for _, domain := range strings.Split(_options.domain, ",") {
	// err := InspectCdnLogTraffic(ctx, domain, _options.begin, _options.end, cdn, _options.mode, trafficrepo)
	// if err != nil {
	// 	util.Fatal(err)
	// }
	// }
}
