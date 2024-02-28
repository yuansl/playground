package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/qbox/net-deftones/fscdn.v2/types"
	"github.com/qbox/net-deftones/logger"
)

const LOG_CDN_HOUR_TABLE = "cdn_hour"

type cdnlogstorage struct {
	hbase HBase
	table string
}

func (logstore *cdnlogstorage) rowkeyof(cdn types.CDNProvider, timestamp time.Time) string {
	key := string(cdn) + ":" + timestamp.Format("2006010215")
	md5sum := md5.Sum([]byte(key))
	return hex.EncodeToString(md5sum[:])[:2] + ":" + key
}

func (logstore *cdnlogstorage) Stat(ctx context.Context, cdn types.CDNProvider, start, end time.Time) ([]CdnLogStat, error) {
	var domains []CdnLogStat

	for hour := start; hour.Before(end); hour = hour.Add(1 * time.Hour) {
		cells, err := logstore.hbase.Get(ctx, logstore.table, logstore.rowkeyof(cdn, hour))
		if err != nil {
			return nil, fmt.Errorf("logstore.hbase.Get: %w", err)
		}
		for _, cell := range cells {
			logger.FromContext(ctx).Infof("cell: %+v\n", cell)
			cf_col := strings.Split(cell.Columnfamily, ":")
			domains = append(domains, CdnLogStat{Domain: cf_col[1]})
		}
	}
	return domains, nil
}

var _ CdnLogRepository = (*cdnlogstorage)(nil)

func NewCdnLogRepository(hbase HBase) *cdnlogstorage {
	return &cdnlogstorage{hbase: hbase, table: LOG_CDN_HOUR_TABLE}
}
