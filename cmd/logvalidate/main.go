package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/qbox/net-deftones/logger"
	"github.com/qbox/net-deftones/util"
	"golang.org/x/sync/errgroup"

	"github.com/yuansl/playground/clients/logetl"
	"github.com/yuansl/playground/clients/loglinks"
)

var (
	ErrBadRequest = errors.New("etl: bad request")
	ErrNotFound   = errors.New("log/list: not found")
)

var (
	concurrency int
	domains     string
	begin       time.Time
	end         time.Time
	granularity string
	force       bool
	sizeMin     int
)

func parseCmdArgs() {
	flag.IntVar(&concurrency, "c", runtime.NumCPU(), "concurrency")
	flag.StringVar(&domains, "domains", "", "specify cdn domain(s), sperated by comma")
	flag.TextVar(&begin, "begin", time.Time{}, "specify begin time")
	flag.TextVar(&end, "end", time.Time{}, "specify end time")
	flag.StringVar(&granularity, "g", "hour", "specify granularity (5min|hour)")
	flag.BoolVar(&force, "force", false, "do whatever necessary if true")
	flag.IntVar(&sizeMin, "size", _FILE_SIZE_MIN, "file size min(water mark)")
	flag.Parse()
}

func inspectEtlTask(ctx context.Context, id string, etlcli *logetl.Client) error {
	r, err := etlcli.GetEtlTasks(&logetl.EtlTaskRequest{Id: id})
	if err != nil {
		return fmt.Errorf("etlTasks: %v", err)
	}

	logger.FromContext(ctx).Infof("result: %+v\n", r)

	return nil
}

func tryEtl(ctx context.Context, domain string, datetime time.Time, etlcli *logetl.Client) error {
	var taskid string

	err := util.WithRetry(ctx, func() error {
		tasks, err := etlcli.SendEtlRetryRequest(ctx, &logetl.EtlRetryRequest{
			Cdn:     "all",
			Domains: []string{domain},
			Start:   datetime,
			End:     datetime.Add(1 * time.Hour),
			Force:   true,
			Manual:  true,
		})
		if err != nil {
			switch {
			case errors.Is(err, logetl.ErrInvalid):
				err = errors.Join(err, context.Canceled)
			default:
			}
			return err
		}
		if len(tasks) > 0 {
			taskid = tasks[0]
		}
		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return nil
		default:
			return err
		}
	}
	return inspectEtlTask(ctx, taskid, etlcli)
}

func fetchDomainLogLinks(domain string, day time.Time, g Granularity, client *loglinks.Client) ([]loglinks.LogLink, error) {
	r, err := client.ListLogLinks(&loglinks.LogListRequest{
		Domains: domain,
		Day:     day,
	})
	if err != nil {
		return nil, err
	}
	if len(r.Result) == 0 {
		return nil, ErrNotFound
	}

	links := r.Result[domain]

	for i := 0; i < len(links); i++ {
		_link := &links[i]
		slices := strings.Split(_link.Name, "_")

		if len(slices) < 3 {
			panic("BUG: unexpected log link name's format:" + _link.Name)
		}

		switch g {
		case Granularity5min:
			_link.Timestamp, err = time.ParseInLocation("2006-01-02-15-04", slices[1][:16], time.Local)
		case Granularity1hour:
			fallthrough
		default:
			_link.Timestamp, err = time.ParseInLocation("2006-01-02-15", slices[1][:13], time.Local)
		}
		if err != nil {
			return nil, fmt.Errorf("time.Parse(%s): %v", slices[1], err)
		}
	}

	return links, nil
}

const (
	_NR_LOG_LINKS_G_hour_PER_DAY = 24
	_NR_LOG_LINKS_G_5min_PER_DAY = 288
)

func NumLogLinksBy(g Granularity) int {
	switch g {
	case Granularity5min:
		return _NR_LOG_LINKS_G_5min_PER_DAY
	case Granularity1hour:
		fallthrough
	default:
		return _NR_LOG_LINKS_G_hour_PER_DAY
	}
}

const _FILE_SIZE_MIN = 1 << 10

type LogLink = loglinks.LogLink

func doValidateDomainLogs(ctx context.Context, domain string, day, end time.Time, g Granularity, fileSizeMin int, client *loglinks.Client, etlcli *logetl.Client) {
	links, err := fetchDomainLogLinks(domain, day, g, client)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			return
		default:
			fatal("fetchDomainLogLinks:", err)
		}
	}

	filterout := func(links []LogLink) bool {
		for _, link := range links {
			if link.Size < int64(fileSizeMin) {
				return false
			}
		}
		return true
	}

	if len(links) == NumLogLinksBy(g) && filterout(links) {
		return
	}
	log := logger.FromContext(ctx)

	log.Printf("got %d links of domain %s at day %v\n", len(links), domain, day.Format(time.DateOnly))

	perHourLogs := make(map[time.Time]*LogLink)

	for i := 0; i < len(links); i++ {
		perHourLogs[links[i].Timestamp] = &links[i]
	}

	step := 1 * time.Hour
	if g == Granularity5min {
		step = 5 * time.Minute
	}
	for datetime := day; datetime.Before(end); datetime = datetime.Add(step) {
		if link, exists := perHourLogs[datetime]; !exists || link.Size < int64(fileSizeMin) {
			log.Printf("WARNING: missed domain %s's log at %v\n", domain, datetime)

			if force {
				err = tryEtl(ctx, domain, datetime, etlcli)
				if err != nil {
					switch {
					case errors.Is(err, ErrBadRequest):
						log.Printf("etl task for domain %s (%v) failed: %v\n", domain, datetime, err)
						continue
					default:
						fatal("tryEtl error:", err)
					}
				}
			}
		}
	}
}

func validateDomainsLogs(ctx context.Context, domains []string, begin, end time.Time, g Granularity, fileSizeMin int, loglinkcli *loglinks.Client, etlcli *logetl.Client) {
	var wg sync.WaitGroup
	var climit = make(chan struct{}, concurrency)

	for day := begin; day.Before(end); day = day.Add(24 * time.Hour) {
		for _, domain := range domains {
			climit <- struct{}{}
			wg.Add(1)

			go func(domain string, day time.Time) {
				defer func() {
					<-climit
					wg.Done()
				}()

				doValidateDomainLogs(ctx, domain, day, end, g, fileSizeMin, loglinkcli, etlcli)
			}(domain, day)
		}
	}
	wg.Wait()
}

func main() {
	parseCmdArgs()

	// validateDomainsLogs(context.TODO(), strings.Split(domains, ","), begin, end, GranularityOf(granularity), sizeMin, loglinks.NewClient(), logetl.NewClient())

	etlcli := logetl.NewClient()
	ctx := logger.NewContext(context.Background(), logger.New())
	wg, ctx := errgroup.WithContext(ctx)
	wg.SetLimit(concurrency)

	for _, d := range strings.Split(domains, ",") {
		if d == "" || d == "domain" {
			continue
		}

		for datetime := begin; datetime.Before(end); datetime = datetime.AddDate(0, 0, +1) {
			_datetime := datetime
			domain := d
			wg.Go(func() error {
				return tryEtl(logger.NewContext(ctx, logger.New()), domain, _datetime, etlcli)
			})
		}
	}
	if err := wg.Wait(); err != nil {
		fatal(err)
	}
	fmt.Printf("DONE\n")
}
