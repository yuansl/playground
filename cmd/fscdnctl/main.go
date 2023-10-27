package main

import (
	"context"
	"flag"
	"os"
	"strings"
	"time"

	"github.com/qbox/net-deftones/fusionrobot/fsrobot/stat"
	"github.com/qbox/net-deftones/fusionrobot/fsrobot/stat/fscdn"
	"github.com/qbox/net-deftones/logger"
	"github.com/qbox/net-deftones/staging/qiniu.com/cdnapi/fusion.v2/fusion"
	"github.com/qbox/net-deftones/util"
)

func init() {
	os.Setenv("TZ", "Asia/Shanghai")
}

var (
	_endpoint string
	_cdn      string
	_start    time.Time = time.Date(2023, 10, 24, 0, 0, 0, 0, time.Local)
	_end      time.Time = _start.Add(24 * time.Hour)
	_domains  string
)

func parseCmdArgs() {
	flag.StringVar(&_endpoint, "endpoint", "xs983:18131", "specify address of the fscdn grpc server")
	flag.StringVar(&_cdn, "cdn", "cloudflaren", "specify a cdn name")
	flag.TextVar(&_start, "start", _start, "specify start time (in RFC3339)")
	flag.TextVar(&_end, "end", _end, "specify end time (in RFC3339)")
	flag.StringVar(&_domains, "domains", "subtitles.netpop.app", "domains (seperated by comma)")
	flag.Parse()
}

func run(ctx context.Context, srv stat.CdnStatService) {
	log := logger.FromContext(ctx)

	{
		res, err := srv.GetCdnDomains(ctx, fusion.CDNProvider(_cdn))
		if err != nil {
			util.Fatal(err)
		}
		log.Infof("GetCdnDomsin Response: %+v\n", res)
	}

	{
		res, err := srv.GetCdnDomainsBandwidth(ctx, fusion.CDNProvider(_cdn), strings.Split(_domains, ","), _start, _end, fusion.Granularity5Min, false)
		if err != nil {
			util.Fatal(err)
		}
		log.Infof("GetCdnDomainsBandwidth res:%+v\n", res)
	}
}

func main() {
	parseCmdArgs()

	srv, err := fscdn.NewCdnStatService(_endpoint)
	if err != nil {
		util.Fatal(err)
	}
	ctx := logger.WithContext(context.TODO(), logger.New())

	run(ctx, srv)
}
