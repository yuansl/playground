package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"time"

	"github.com/qbox/net-deftones/fscdn.v2"
	"github.com/qbox/net-deftones/fscdn.v2/types"
	statfscdn "github.com/qbox/net-deftones/fusionrobot/fsrobot/stat/fscdn"
	"github.com/qbox/net-deftones/logger"
	"github.com/qbox/net-deftones/util"
)

var (
	_domain         string
	_begin          time.Time
	_end            time.Time
	_cdn            string
	_granularity    string
	_grpcServerAddr string
)

func parseCmdArgs() {
	flag.StringVar(&_domain, "domain", "www.example.com", "specify cdn domain")
	flag.TextVar(&_begin, "begin", time.Time{}, "specify begin time (in RFC3339 format)")
	flag.TextVar(&_end, "end", time.Time{}, "specify end time (in RFC3339 format)")
	flag.StringVar(&_cdn, "cdn", "qiniucdn", "specify cdn provider name")
	flag.StringVar(&_granularity, "g", "5min", "specify granularity")
	flag.StringVar(&_grpcServerAddr, "addr", "localhost:80", "specify fscdn grpc server address")
	flag.Parse()
}

func cdnStatServiceDemo(ctx context.Context, start, end time.Time, domains []string, cdn string, g types.Granularity) {
	srv, err := statfscdn.NewCdnStatService(_grpcServerAddr)
	if err != nil {
		util.Fatal(err)
	}
	switch g {
	case types.Granularity1Min, types.Granularity5Min:
		// do nothing
	default:
		util.Fatal("Invalid argument: unknown granularity '%s'\n", g)
	}

	res, err := srv.GetCdnDomainsBandwidth(ctx, types.CDNProvider(cdn), domains, start, end, g, false)
	if err != nil {
		switch {
		case errors.Is(err, fscdn.ErrInvalidDomain):
			fmt.Printf("fscdn says its a invalid domain: %v\n", err)
		default:
			util.Fatal(err)
		}
	}
	for _, it := range res {
		fmt.Printf("Metric: '%v', domain='%s', region='%v', timeseries=%v\n", it.DataType, it.Domain, it.GeoCover, it.BandWidth)
	}
}

func main() {
	parseCmdArgs()

	ctx := logger.NewContext(context.TODO(), logger.New())

	cdnStatServiceDemo(ctx, _begin, _end, []string{_domain}, _cdn, _granularity)
}
