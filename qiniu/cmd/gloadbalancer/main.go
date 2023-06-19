package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/qbox/net-deftones/cdnapi.v2/googlecloud"
	"github.com/qbox/net-deftones/logger"
)

const PROJECT_ID = "qiniu-pro"

var (
	CAFile    string
	endpoint  string
	projectId string
	start     time.Time
	end       time.Time
)

func parseCmdArgs() {
	flag.StringVar(&endpoint, "endpoint", "defy-googleapisproxy.qiniuapi.com:443", "specify the endpoint of googleapiserver proxy")
	flag.StringVar(&CAFile, "CAFile", "/etc/ssl/certs/ca-certificates.crt", "specify CA file")
	flag.StringVar(&projectId, "project", PROJECT_ID, "specify the project-id of google cloud service")
	flag.TextVar(&start, "start", time.Date(2023, 3, 1, 0, 0, 0, 0, time.Local), "specify the start date in RFC3339 format")
	flag.TextVar(&end, "end", time.Date(2023, 3, 2, 0, 0, 0, 0, time.Local), "specify the end date in RFC3339 format")
	flag.Parse()
}

func fatal(v ...any) {
	fmt.Fprintln(os.Stderr, "fatal error: ", v)
	os.Exit(1)
}

func main() {
	parseCmdArgs()

	lb, err := googlecloud.NewLoadbalancingService(&googlecloud.GcloudOptions{
		ProjectId: projectId,
		Host:      endpoint,
		EnableTLS: true,
		CAFile:    CAFile,
	})
	if err != nil {
		fatal(err)
	}
	ctx := logger.NewContext(context.TODO(), logger.New())

	urlmaps, err := lb.ListLoadBalancerUrlMapNames(ctx)
	if err != nil {
		fatal(err)
	}
	stats, err := lb.FetchTimeSeriesStatOf(ctx, urlmaps, start, end)
	if err != nil {
		fatal("lb.FetchTimeSeriesStatOf error:", err)
	}

	for _, stat := range stats {
		fmt.Printf("stat: +%v\n", stat)
	}
}
