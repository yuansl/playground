package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/qbox/net-deftones/fscdn.v2/types"
	"github.com/qbox/net-deftones/fusionrobot"
)

type Cell struct {
	Row       string
	Column    string
	Timestamp time.Time
	Value     any
}

type Hbase interface {
	// Get row or cell contents. pass table name, rowkey, and optional column family
	Get(_ context.Context, table, rowkey string, columnfamily ...string) ([]Cell, error)
	DeleteAll(ctx context.Context, table, rowkey string) error
}

type hbaseStorage struct {
	hbase Hbase
}

func rowkeyof(domain string, cdn types.CDNProvider, region fusionrobot.Region, hour time.Time) string {
	return fmt.Sprintf("%[1]s|%[2]s|%02[3]d|%[4]s|%[5]s",
		hour.Local().Format(time.DateOnly), domain, hour.Hour(), cdn, region.String())
}

func (store *hbaseStorage) Stat(ctx context.Context, domain string, cdn types.CDNProvider, region fusionrobot.Region, begin, end time.Time) ([]Stat, error) {
	var stats []Stat

	for hour := begin; hour.Before(end); hour = hour.Add(1 * time.Hour) {
		rowkey := rowkeyof(domain, cdn, region, hour)
		cells, err := store.hbase.Get(ctx, _HBASE_TABLENAME, rowkey, "0", "1", "2", "3", "4", "5")
		if err != nil {
			return nil, fmt.Errorf("hbase.get: %w", err)
		}
		for _, cell := range cells {
			bytes, err := strconv.Atoi(cell.Value.(string))
			if err != nil {
				return nil, fmt.Errorf("ill-formated integer of '%v'", cell.Value)
			}
			stats = append(stats, Stat{
				Domain:    domain,
				Cdn:       cdn,
				Region:    region,
				Timestamp: cell.Timestamp,
				Value:     bytes,
			})
		}
	}
	return stats, nil
}

func NewHbaseStorage(hbase Hbase) *hbaseStorage {
	return &hbaseStorage{hbase: hbase}
}
