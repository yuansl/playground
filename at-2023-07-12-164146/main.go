// -*- mode:go; mode:go-playground -*-
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	ErrBadRequest = errors.New("etl: bad request")
	ErrNotFound   = errors.New("log/list: not found")
)

var (
	force       bool
	concurrency int
	domains     string
	begin       time.Time
	end         time.Time
	granularity string
)

func parseCmdArgs() {
	flag.IntVar(&concurrency, "c", runtime.NumCPU(), "concurrency")
	flag.BoolVar(&force, "force", false, "try do whatever nessary if true")
	flag.StringVar(&domains, "domains", "", "specify cdn domain(s), sperated by comma")
	flag.TextVar(&begin, "begin", time.Time{}, "specify begin time")
	flag.TextVar(&end, "end", time.Time{}, "specify end time")
	flag.StringVar(&granularity, "g", "hour", "specify granularity (5min|hour)")
	flag.Parse()
}

//go:generate stringer -type LogType -linecomment
type LogType int

const (
	CdnLogType LogType = iota // cdn
	SrcLogType                // src
)

//go:generate stringer -type Granularity -linecomment
type Granularity int

const (
	Granularity1hour Granularity = iota // hour
	Granularity5min                     // 5min
)

func GranularityOf(name string) Granularity {
	switch name {
	case "5min":
		return Granularity5min
	case "hour":
		return Granularity1hour
	default:
		panic("BUG: unknown granularity:" + name)
	}
}

func inspectEtlTask(id string) error {
	r, err := ListEtlTasks(&EtlTaskRequest{Id: id})
	if err != nil {
		return fmt.Errorf("etlTasks: %v", err)
	}

	log.Printf("result: %+v\n", r)

	return nil
}

func tryEtl(domain string, datetime time.Time) error {
	r, err := CreateEtlTask(&EtlRequest{Cdn: "all", Domains: domain, Hour: datetime, Type: BandwidthData, Force: true})
	if err != nil {
		return err
	}
	return inspectEtlTask(r.TaskId)
}

func fetchDomainLogLinks(domain string, day time.Time, g Granularity) ([]LogLink, error) {
	r, err := ListLogLinks(&LogLinkRequest{
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

func NumLogLinks(g Granularity) int {
	switch g {
	case Granularity5min:
		return _NR_LOG_LINKS_G_5min_PER_DAY
	case Granularity1hour:
		fallthrough
	default:
		return _NR_LOG_LINKS_G_hour_PER_DAY
	}
}

const _FILE_SIZE_MIN = 1 << 20

func doValidateDomainLogs(domain string, day, end time.Time, g Granularity) {
	links, err := fetchDomainLogLinks(domain, day, g)
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
			if link.Size < _FILE_SIZE_MIN {
				return false
			}
		}
		return true
	}

	if len(links) == NumLogLinks(g) && filterout(links) {
		return
	}

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
		if link, exists := perHourLogs[datetime]; !exists || link.Size < _FILE_SIZE_MIN {
			log.Printf("WARNING: missed domain %s's log at %v\n", domain, datetime)

			if force {
				err = tryEtl(domain, datetime)
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

func validateDomainsLogs(domains []string, begin, end time.Time, g Granularity) {
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

				doValidateDomainLogs(domain, day, end, g)
			}(domain, day)
		}
	}
	wg.Wait()
}

func main() {
	parseCmdArgs()

	validateDomainsLogs(strings.Split(domains, ","), begin, end, GranularityOf(granularity))
}
