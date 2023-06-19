package main

import (
	"context"
	"flag"
	"fmt"
	"runtime"
	"time"

	"golang.org/x/sync/errgroup"
)

var (
	from      time.Time
	to        time.Time
	uidfile   string
	mysqladdr string
	all       bool
)

func parseCmdArgs() {
	flag.TextVar(&from, "from", time.Now(), "specify the start time of the task in the format RFC3339")
	flag.TextVar(&to, "to", time.Now(), "specify the end time of the task in the format RFC3339")
	flag.StringVar(&uidfile, "file", "", "uids file(csv)")
	flag.StringVar(&mysqladdr, "mysqladdr", "traffic_admin:kdK26Ws824Q9ivvfao2ns@(xs614:3359,xs598:3359,jjh873:3359)/traffic?parseTime=true&loc=Local", "specify the address of the traffic service's database(mysql dsn)")
	flag.BoolVar(&all, "all", false, "stat traffic of all of uids")
	flag.Parse()
}

func statDynReqcountOf(ctx context.Context, uids []uint32, from, to time.Time, trafficSrv TrafficService) int64 {
	trafficq := make(chan []TrafficStat, runtime.NumCPU())

	go func() {
		defer close(trafficq)

		egroup, ctx := errgroup.WithContext(ctx)

		egroup.SetLimit(runtime.NumCPU())

		for _, uid := range uids {
			_uid := uid
			egroup.Go(func() error {
				traffics, err := trafficSrv.GetCdnTrafficOf(ctx, from, to, DataTypeDynReqcount, _uid)
				if err != nil {
					return fmt.Errorf("trafficSrv.GetCdnTrafficOf: %v", err)
				}

				trafficq <- traffics

				return nil
			})
		}
		egroup.Wait()
	}()

	var sum int64
	for traffics := range trafficq {
		for _, traffic := range traffics {
			sum += Sum(traffic.Timeseries...)
		}
	}
	return sum
}

func statDynReqcount(ctx context.Context, from, to time.Time, trafficSrv TrafficService, uids ...uint32) int64 {
	if len(uids) > 0 {
		return statDynReqcountOf(ctx, uids, from, to, trafficSrv)
	}

	traffics, err := trafficSrv.GetAllOfCdnTraffics(ctx, from, to, DataTypeDynReqcount)
	if err != nil {
		fatal("trafficSrv.GetAllOfCdnTraffics error:", err)
	}

	var sum int64
	for _, traffic := range traffics {
		sum += Sum(traffic.Timeseries...)
	}
	return sum
}

func main() {
	parseCmdArgs()

	trafficSrv := NewTrafficService(WithDSN(mysqladdr))
	uids := []uint32{}
	if !all {
		uids = loadUIDsFrom(uidfile)
		uids = deduplicateOf(uids)
		fmt.Printf("load %d uids in total\n", len(uids))
	}

	sum := statDynReqcount(context.TODO(), from, to, trafficSrv, uids...)

	fmt.Printf("year:%v, sum=%d\n", from.Format("2006"), sum)
}
