package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/qbox/net-deftones/googleapisproxy"
	"github.com/qbox/net-deftones/logger"
)

func fatal(v ...any) {
	fmt.Fprintln(os.Stderr, "fatal error: ", v)
	os.Exit(1)
}

const _PROJECT_ID = "qiniu-pro"

func main() {
	req := googleapisproxy.GcloudClientOptions{}
	client, err := googleapisproxy.NewGcloudClient("localhost:18131", &req)
	if err != nil {
		fatal(err)
	}
	ctx := logger.NewContext(context.TODO(), logger.New())

	res, err := client.ListUrlMaps(ctx, &googleapisproxy.ListUrlMapsRequest{
		ProjectId: _PROJECT_ID,
	})
	if err != nil {
		fatal("client.ListUrlMaps error:", err)
	}
	fmt.Println("urlmaps:", res.Urlmaps)

	start := time.Date(2023, 3, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 0, +1)

	client.FetchMetricStatOfCloudCdn(ctx, &googleapisproxy.TimeseriesStatRequest{
		StartUnixtime: start.Unix(),
		EndUnixtime:   end.Unix(),
		ProjectId:     _PROJECT_ID,
	})
}
