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
	"github.com/qbox/net-deftones/util"
)

const (
	LOG_CDN_HOUR_TABLE       = "cdn_hour"
	LOG_DOMAIN_HOUR_TABLE    = "domain_hour"
	LOG_CDN_LOG_TABLE_PREFIX = "cdnLog_"
)

type Cell struct {
	Row          string
	Columnfamily string
	Timestamp    time.Time
	Value        []byte
}

type HBase interface {
	// Get row or cell contents. pass table name, rowkey, and optional column family
	Get(ctx context.Context, table, key string, col ...string) ([]Cell, error)
	Delete(ctx context.Context, table, key string, col ...string) error
	Scan(ctx context.Context, table string, col ...string) ([]Cell, error)
}

type cdnlogstorage struct {
	hbase           HBase
	cdnHourTable    string
	domainHourTable string
	cdnLogTable     string
}

func (logmeta *cdnlogstorage) FetchDomainsOf(ctx context.Context, cdn types.CDNProvider, start, end time.Time) ([]string, error) {
	var domains []string

	for hour := start; hour.Before(end); hour = hour.Add(1 * time.Hour) {
		cells, err := logmeta.hbase.Get(ctx, logmeta.cdnHourTable, logmeta.RowkeyofCdnHour(cdn, hour))
		if err != nil {
			return nil, fmt.Errorf("logstore.hbase.Get: %w", err)
		}
		for _, cell := range cells {
			logger.FromContext(ctx).Infof("cell: %+v\n", cell)
			cf_col := strings.SplitN(cell.Columnfamily, ":", 2)

			util.Assert(len(cf_col) == 2, "The cell's columnfamily mismatch in table domain_hour")

			domains = append(domains, cf_col[1])
		}
	}
	return domains, nil
}

func (logmeta *cdnlogstorage) FetchCdnProvidersOf(ctx context.Context, domain string, start, end time.Time) ([]string, error) {
	var cdns []string

	for hour := start; hour.Before(end); hour = hour.Add(1 * time.Hour) {
		cells, err := logmeta.hbase.Get(ctx, logmeta.domainHourTable, logmeta.RowkeyofDomainHour(domain, hour))
		if err != nil {
			return nil, fmt.Errorf("logstore.hbase.Get: %w", err)
		}
		for _, cell := range cells {
			logger.FromContext(ctx).Infof("cell: %+v\n", cell)
			cf_col := strings.SplitN(cell.Columnfamily, ":", 2)

			util.Assert(len(cf_col) == 2, "The cell's columnfamily mismatch in table domain_hour")

			cdns = append(cdns, cf_col[1])
		}
	}
	return cdns, nil
}

func (logmeta *cdnlogstorage) cdnLogTableName(datetime time.Time) string {
	return LOG_CDN_LOG_TABLE_PREFIX + datetime.Format("20060102")
}

func (logmeta *cdnlogstorage) FetchCdnTrafficOf(ctx context.Context, domain, cdn string, start, end time.Time) ([]string, error) {
	var cdns []string

	for hour := start; hour.Before(end); hour = hour.Add(1 * time.Hour) {
		cells, err := logmeta.hbase.Scan(ctx, logmeta.cdnLogTableName(hour))
		if err != nil {
			return nil, fmt.Errorf("logstore.hbase.Get: %w", err)
		}
		for _, cell := range cells {
			logger.FromContext(ctx).Infof("cell: %+v\n", cell)
			cf_col := strings.SplitN(cell.Columnfamily, ":", 2)

			util.Assert(len(cf_col) == 2, "The cell's columnfamily mismatch in table domain_hour")

			cdns = append(cdns, cf_col[1])
		}
	}
	return cdns, nil
}

func (logmeta *cdnlogstorage) RowkeyofCdnHour(cdn types.CDNProvider, timestamp time.Time) string {
	key := string(cdn) + ":" + timestamp.Format("2006010215")
	md5sum := md5.Sum([]byte(key))
	return hex.EncodeToString(md5sum[:])[:2] + ":" + key
}

func (logmeta *cdnlogstorage) RowkeyofDomainHour(domain string, timestamp time.Time) string {
	key := domain + ":" + timestamp.Format("2006010215")
	md5sum := md5.Sum([]byte(key))
	return hex.EncodeToString(md5sum[:])[:2] + ":" + key
}

var _ CdnLogRepository = (*cdnlogstorage)(nil)

func NewCdnLogRepository(hbase HBase) *cdnlogstorage {
	return &cdnlogstorage{
		hbase:           hbase,
		cdnHourTable:    LOG_CDN_HOUR_TABLE,
		domainHourTable: LOG_DOMAIN_HOUR_TABLE,
	}
}
