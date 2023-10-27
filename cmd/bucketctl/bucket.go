package main

import (
	"context"
	"fmt"
	"time"

	"github.com/yuansl/playground/clients/sophon"
	"github.com/yuansl/playground/cmd/bucketctl/internal/bucket/xsophon"
)

func TestGetBucket(domain string, time time.Time) {
	bucketsrv := xsophon.NewBucketService(sophon.NewClient())
	bucket, err := bucketsrv.GetBucket(context.TODO(), domain, time)
	if err != nil {
		fatal(err)
	}

	fmt.Printf("bucket of domain %s: %q\n", globalOptions.domain, bucket)
}
