package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/qbox/net-deftones/fscdn.v2/types"
	"github.com/qbox/net-deftones/fusionrobot"
	"golang.org/x/sync/errgroup"
)

const _LOG_TRAFFIC_TABLE = "unify_flux"

type cdnlogtraffic struct {
	hbase HBase
}

func (store *cdnlogtraffic) rowkeyof(domain string, cdn types.CDNProvider, region fusionrobot.Region, hour time.Time) string {
	return fmt.Sprintf("%[1]s|%[2]s|%02[3]d|%[4]s|%[5]s",
		hour.Local().Format(time.DateOnly), domain, hour.Hour(), cdn, region.String())
}

func (store *cdnlogtraffic) DeleteAll(ctx context.Context, domain string, cdn types.CDNProvider, region fusionrobot.Region, begin, end time.Time) error {
	wg, ctx := errgroup.WithContext(ctx)

	for hour := begin; hour.Before(end); hour = hour.Add(1 * time.Hour) {
		for i := range 12 {
			wg.Go(func() error {
				col := "cf:" + strconv.Itoa(i)
				err := store.hbase.Delete(ctx, _LOG_TRAFFIC_TABLE, store.rowkeyof(domain, cdn, region, hour), col)
				if err != nil {
					return fmt.Errorf("hbase.Delete: %w", err)
				}
				return nil
			})
		}
	}
	return wg.Wait()
}

func (store *cdnlogtraffic) Stat(ctx context.Context, domain string, cdn types.CDNProvider, region fusionrobot.Region, begin, end time.Time) ([]Stat, error) {
	var statch = make(chan []Stat)

	go func() {
		defer close(statch)

		wg, ctx := errgroup.WithContext(ctx)

		for hour := begin; hour.Before(end); hour = hour.Add(1 * time.Hour) {
			wg.Go(func() error {
				rowkey := store.rowkeyof(domain, cdn, region, hour)
				cells, err := store.hbase.Get(ctx, _LOG_TRAFFIC_TABLE, rowkey, "cf:0", "cf:1", "cf:2")
				if err != nil {
					return fmt.Errorf("hbase.get: %w", err)
				}

				var stats []Stat

				for _, cell := range cells {
					bytes, err := strconv.Atoi(string(cell.Value))
					if err != nil {
						return fmt.Errorf("ill-formated integer of '%v'", cell.Value)
					}
					stats = append(stats, Stat{
						Domain:    domain,
						Cdn:       cdn,
						Region:    region,
						Timestamp: cell.Timestamp,
						Value:     bytes,
					})
				}
				statch <- stats
				return nil
			})
		}
		wg.Wait()
	}()
	var stats []Stat

	for stat := range statch {
		stats = append(stats, stat...)
	}
	return stats, nil
}

func NewCdnLogTraffic(hbase HBase) *cdnlogtraffic {
	return &cdnlogtraffic{hbase: hbase}
}
