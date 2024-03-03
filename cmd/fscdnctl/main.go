package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"time"

	"github.com/qbox/net-deftones/fscdn.v2/types"
	statfscdn "github.com/qbox/net-deftones/fusionrobot/fsrobot/stat/fscdn"
	"github.com/qbox/net-deftones/logger"
	"github.com/qbox/net-deftones/util"
)

var options struct {
	domain      string
	begin       time.Time
	end         time.Time
	cdn         string
	granularity string
	grpcServer  string
}

func parseOptions() {
	flag.StringVar(&options.domain, "domain", "www.example.com", "specify cdn domain")
	flag.TextVar(&options.begin, "begin", time.Time{}, "specify begin time (in RFC3339 format)")
	flag.TextVar(&options.end, "end", time.Time{}, "specify end time (in RFC3339 format)")
	flag.StringVar(&options.cdn, "cdn", "qiniucdn", "specify cdn provider name")
	flag.StringVar(&options.granularity, "g", "5min", "specify granularity")
	flag.StringVar(&options.grpcServer, "addr", "localhost:80", "specify fscdn grpc server address")
	flag.Parse()
}

func cdnStatServiceDemo(ctx context.Context, start, end time.Time, domains []string, cdn string, g types.Granularity) {
	srv, err := statfscdn.NewCdnStatService(options.grpcServer)
	if err != nil {
		util.Fatal(err)
	}
	switch g {
	case types.Granularity1Min, types.Granularity5Min:
		// do nothing
	default:
		util.Fatal("Invalid argument: unknown granularity '%s'\n", g)
	}

	res, err := srv.GetCdnDomainsBandwidth(ctx, types.CDNProvider(cdn), domains, start, end, g, false, false)
	if err != nil {
		switch {
		case errors.Is(err, types.ErrInvalidDomain):
			fmt.Printf("fscdn says its a invalid domain: %v\n", err)
		default:
			util.Fatal(err)
		}
	}
	for _, it := range res {
		fmt.Printf("Metric: '%v', domain='%s', region='%v', timeseries=%v\n", it.DataType, it.Domain, it.Region, it.Timeseries)
	}
}

func main() {
	parseOptions()

	ctx := logger.NewContext(context.TODO(), logger.New())

	cdnStatServiceDemo(ctx, options.begin, options.end, []string{options.domain}, options.cdn, options.granularity)
}
