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
	"time"

	"github.com/qbox/net-deftones/fscdn.v2/types"
	"github.com/qbox/net-deftones/fusionrobot"
	"github.com/qbox/net-deftones/logger"
	"github.com/yuansl/playground/util"
	"golang.org/x/sync/errgroup"
)

type Stat struct {
	Domain    string
	Cdn       types.CDNProvider
	Region    fusionrobot.Region
	Timestamp time.Time
	Value     int
}

type CdnTrafficStorage interface {
	// Delete(ctx context.Context, domain string, cdn types.CDNProvider, region fusionrobot.Region, start, end time.Time) error
	Stat(ctx context.Context, domain string, cdn types.CDNProvider, region fusionrobot.Region, start, end time.Time) ([]Stat, error)
}

func deleteForeignRegionTraffics(ctx context.Context, domain string, begin, end time.Time, cdn types.CDNProvider, store CdnTrafficStorage) error {
	wg, ctx := errgroup.WithContext(ctx)
	log := logger.FromContext(ctx)

	for r := fusionrobot.RegionAMEU; r <= fusionrobot.RegionOC; r++ {
		_region := r
		wg.Go(func() error {
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			stats, err := store.Stat(ctx, domain, cdn, _region, begin, end)
			if err != nil {
				return fmt.Errorf("store.Stat: %w", err)
			}
			for _, p := range stats {
				log.Infof("traffic stat: %+v\n", p)
			}
			// hbase.DeleteAll(context.TODO(), _HBASE_TABLENAME, rowkey)
			return nil
		})
	}
	return wg.Wait()
}

var _options struct {
	domain    string
	begin     time.Time
	end       time.Time
	cdn       string
	hbaseaddr string
}

func parseOptions() {
	flag.StringVar(&_options.domain, "domains", "file.yalla.live", "specify domains(sperate by comma)")
	flag.TextVar(&_options.begin, "begin", time.Time{}, "specify begin time in RFC3339 format")
	flag.TextVar(&_options.end, "end", time.Time{}, "specify end time in RFC3339 format")
	flag.StringVar(&_options.cdn, "cdn", string(types.CDNProviderBaishanyun), "specify cdn provider")
	flag.StringVar(&_options.hbaseaddr, "hbase", "localhost", "specify hbase address")
	flag.Parse()
}

func main() {
	parseOptions()

	cdn := types.CDNProvider(_options.cdn)
	hbase := OpenHbase(_options.hbaseaddr)
	store := NewHbaseStorage(hbase)
	ctx := logger.NewContext(context.Background(), logger.New())

	defer hbase.Close()

	for _, domain := range strings.Split(_options.domain, ",") {
		err := deleteForeignRegionTraffics(ctx, domain, _options.begin, _options.end, cdn, store)
		if err != nil {
			util.Fatal(err)
		}
	}
}
