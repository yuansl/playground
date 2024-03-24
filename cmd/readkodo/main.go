package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/qbox/net-deftones/util"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/yuansl/playground/oss"
	"github.com/yuansl/playground/oss/kodo"
	playtrace "github.com/yuansl/playground/trace"
)

const (
	_ACCESS_KEY_DEFAULT  = "557TpseUM8ovpfUhaw8gfa2DQ0104ZScM-BTIcBx"
	_SECRET_KEY_DEFAULT  = "d9xLPyreEG59pR01sRQcFywhm4huL-XEpHHcVa90"
	_KODO_BUCKET_DEFAULT = "fusionlogtest"
)

var (
	_accessKey = _ACCESS_KEY_DEFAULT
	_secretKey = _SECRET_KEY_DEFAULT
	_bucket    = _KODO_BUCKET_DEFAULT
	_prefix    string
	_limit     int
)

var _applicationTracer = playtrace.GetTracerProvider().Tracer("")

func init() {
	if env := os.Getenv("ACCESS_KEY"); env != "" {
		_accessKey = env
	}
	if env := os.Getenv("SECRET_KEY"); env != "" {
		_secretKey = env
	}
}

func parseCmdArgs() {
	flag.StringVar(&_bucket, "bucket", _KODO_BUCKET_DEFAULT, "specify bucket name")
	flag.StringVar(&_prefix, "prefix", "", "specify a bucket key prefix")
	flag.IntVar(&_limit, "limit", 5, "limit number of list")
	flag.Parse()
}

func Download(ctx context.Context, bucket, name string, srv oss.ObjectStorageService) {
	data, err := srv.Download(ctx, bucket, name)
	if err != nil {
		util.Fatal("kodo.Download:", err)
	}

	span := trace.SpanFromContext(ctx)

	if span.IsRecording() {
		span.SetAttributes(attribute.String("name", name))
		span.SetAttributes(attribute.Int("databytes", len(data)))
	}

	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		util.Fatal("gzip.NewReader:", err)
	}
	defer gz.Close()

	fp, err := os.CreateTemp("/tmp/", "*")
	defer fp.Close()
	b := bufio.NewWriter(fp)
	io.Copy(b, gz)
	b.Flush()

	if span.IsRecording() {
		stat, _ := fp.Stat()
		span.SetAttributes(attribute.Int("gzbytes", int(stat.Size())))
	}
}

func listBucketFiles(ctx context.Context, bucket string, srv oss.ObjectStorageService) {}

func main() {
	parseCmdArgs()

	ctx, span := _applicationTracer.Start(context.TODO(), "kodo.List")
	defer span.End()

	kodosrv := kodo.NewStorageService(
		kodo.WithCredential(_accessKey, _secretKey),
		kodo.WithLinkDomain("http://ria8j59xt.hd-bkt.clouddn.com"),
	)
	options := []oss.ListOption{kodo.WithListLimit(_limit)}
	if _prefix != "" {
		options = append(options, kodo.WithListPrefix(_prefix))
	}
	files, err := kodosrv.List(ctx, _bucket, options...)
	if err != nil {
		util.Fatal(err)
	}
	for _, file := range files {
		func() {
			start := time.Now()
			defer func() {
				fmt.Printf("time elapsed for downing file %s: %v\n", file.Name, time.Since(start))
			}()

			ctx, span := _applicationTracer.Start(ctx, "Download")
			defer span.End()

			Download(ctx, _bucket, file.Name, kodosrv)
		}()
	}
}
