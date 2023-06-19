package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/qbox/net-deftones/fusionrobot"
	"github.com/qbox/net-deftones/fusionrobot/dao"
	"github.com/qbox/net-deftones/fusionrobot/dao/mysql"
	"github.com/qbox/net-deftones/logger"
	"github.com/qbox/net-deftones/staging/qiniu.com/cdnapi/fusion.v2/fusion"
)

const (
	_NR_TIMESERIES      = 288
	_BATCH_SIZE_DEFAULT = 10000
)

func fatal(v ...any) {
	fmt.Fprintln(os.Stderr, "fatal error: ", v)
	os.Exit(1)
}

var (
	nrRows   int
	dbAddr   string
	from, to time.Time
)

func parseCmdArgs() {
	flag.IntVar(&nrRows, "n", 100_000, "specify the number of records will be inserted into")
	flag.StringVar(&dbAddr, "db", "127.0.0.1:3306", "specify the database address")
	flag.TextVar(&from, "from", time.Date(2023, 3, 1, 0, 0, 0, 0, time.Local), "specify the start date")
	flag.TextVar(&to, "to", time.Date(2023, 3, 1, 0, 0, 0, 0, time.Local), "specify the end date")
	flag.Parse()
}

func alignas(t time.Time) time.Time {
	x := t.Unix() / 300 * 300
	return time.Unix(x, 0)
}

func generateTimeseries(day time.Time, size int) fusionrobot.Timeseries {
	timeseries := make([]fusionrobot.TrafficPoint, 0, size)
	for j := 0; j < size; j++ {
		timeseries = append(timeseries, fusionrobot.TrafficPoint{
			Timestamp: day.Add(time.Duration(j) * 5 * time.Minute), Value: rand.Int63() % (100 << 40)})
	}
	return timeseries
}

func init() {
	time.Local = time.UTC
}

var (
	regions = []fusionrobot.Region{fusionrobot.RegionChina, fusionrobot.RegionAsia, fusionrobot.RegionSEA, fusionrobot.RegionAmeu, fusionrobot.RegionSA, fusionrobot.RegionOC}

	sources = []fusionrobot.SourceType{fusionrobot.SourceTypeAPI, fusionrobot.SourceTypeLog}

	datatypes = []fusionrobot.DataType{fusionrobot.DataTypeBandwidth, fusionrobot.DataTypeReqCount}
)

func bulkWriteRawTraffics(ctx context.Context, db dao.TrafficDao, batchsize, nTimeseris int) {
	traffics := []*fusionrobot.RawDayTraffic{}
	start := time.Now()
	count := 0

	defer func() {
		logger.FromContext(ctx).Infof("average insert speed: %f\n",
			float64(count)/float64(time.Since(start).Milliseconds()))
	}()

	go func() {
		log := logger.FromContext(ctx)
		for range time.Tick(5 * time.Second) {
			log.Infof("Insert %d rows at now\n", count)
		}
	}()

	for day := from; day.Before(to); day = day.AddDate(0, 0, +1) {
		for _, dtype := range datatypes {
			for _, region := range regions {
				for _, source := range sources {
					for _, it := range []struct {
						bucket  string
						domains []string
					}{
						{bucket: "cdn-bucket", domains: []string{"www.example.com"}},
						{bucket: "mail-bucket", domains: []string{"mail.example.com", "mail2.example.com", "mail3.example.com"}},
						{bucket: "ftp-bucket", domains: []string{"ftp.example.com", "ftp2.example.com", "ftp3.example.com"}},
						{bucket: "git-bucket", domains: []string{"git.example.com", "git2.example.com", "git3.example.com"}},
					} {
						for _, domain := range it.domains {
							traffics = append(traffics, &fusionrobot.RawDayTraffic{
								Domain:     domain,
								Bucket:     it.bucket,
								Region:     region,
								Day:        day,
								CDN:        fusion.CDNProviderCloudflare,
								SourceType: source,
								DataType:   dtype,
								Timeseries: generateTimeseries(day, nTimeseris),
							})

							if len(traffics) >= batchsize {
								err := db.UpsertRawDayTrafficTable(ctx, traffics)
								if err != nil {
									fatal("UpsertRawDayTrafficTable error:", err)
								}

								count += len(traffics)
								if count >= nrRows {
									return
								}

								traffics = traffics[:0]
							}
						}
					}
				}
			}
		}
	}

	if len(traffics) > 0 {
		err := db.UpsertRawDayTrafficTable(ctx, traffics)
		if err != nil {
			fatal("UpsertRawDayTrafficTable error:", err)
		}
		count += len(traffics)
	}
}

func bulkInsertDomainTraffics(ctx context.Context, db dao.TrafficDao, batchsize, nTimeseris int) {
	traffics := []*fusionrobot.DomainDayTraffic{}
	start := time.Now()
	count := 0

	defer func() {
		logger.FromContext(ctx).Infof("average insert speed: %f\n",
			float64(count)/float64(time.Since(start).Milliseconds()))
	}()

	go func() {
		log := logger.FromContext(ctx)
		for range time.Tick(5 * time.Second) {
			log.Infof("Insert %d rows at now\n", count)
		}
	}()

	for day := from; day.Before(to); day = day.AddDate(0, 0, +1) {
		for _, dtype := range datatypes {
			for _, region := range regions {

				for _, it := range []struct {
					bucket  string
					domains []string
				}{
					{bucket: "cdn-bucket", domains: []string{"www.example.com"}},
					{bucket: "mail-bucket", domains: []string{"mail.example.com", "mail2.example.com", "mail3.example.com"}},
					{bucket: "ftp-bucket", domains: []string{"ftp.example.com", "ftp2.example.com", "ftp3.example.com"}},
					{bucket: "git-bucket", domains: []string{"git.example.com", "git2.example.com", "git3.example.com"}},
				} {
					for _, domain := range it.domains {
						traffics = append(traffics, &fusionrobot.DomainDayTraffic{
							Domain:     domain,
							Bucket:     it.bucket,
							Region:     region,
							Day:        day,
							DataType:   dtype,
							Timeseries: generateTimeseries(day, nTimeseris),
						})

						if len(traffics) >= batchsize {
							err := db.UpsertDomainDayTraffic(ctx, traffics)
							if err != nil {
								fatal("UpsertRawDayTrafficTable error:", err)
							}

							count += len(traffics)
							if count >= nrRows {
								return
							}

							traffics = traffics[:0]
						}
					}
				}
			}
		}
	}

	if len(traffics) > 0 {
		err := db.UpsertDomainDayTraffic(ctx, traffics)
		if err != nil {
			fatal("UpsertRawDayTrafficTable error:", err)
		}
		count += len(traffics)
	}
}

func main() {
	parseCmdArgs()

	db, err := mysql.NewPartitionTrafficDao(&mysql.TrafficDaoConf{
		Mysqluri:  "traffic_admin:admin@(" + dbAddr + ")/traffic?parseTime=true&loc=UTC",
		BatchSize: _BATCH_SIZE_DEFAULT,
	})
	if err != nil {
		fatal(err)
	}

	ctx := logger.NewContext(context.TODO(), logger.New())

	bulkInsertDomainTraffics(ctx, db, _BATCH_SIZE_DEFAULT, _NR_TIMESERIES)
}
