package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/yuansl/playground/oss"
	"github.com/yuansl/playground/oss/kodo"
	trace "github.com/yuansl/playground/trace"
	"github.com/yuansl/playground/util"
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

func Download(ctx context.Context, name string, kodosrv oss.ObjectStorageService) {
	data, err := kodosrv.Download(ctx, name)
	if err != nil {
		util.Fatal("kodo.Download:", err)
	}
	fmt.Printf("got %d bytes from kodo: %s\n", len(data), name)

	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		util.Fatal("gzip.NewReader:", err)
	}
	defer gz.Close()

	data, err = io.ReadAll(gz)
	if err != nil {
		if !errors.Is(err, io.EOF) {
			util.Fatal("io.ReadAll(gzip(%s)) error: %v\n", name, err)
		}
	}
	fmt.Printf("got %d from gzip file %s\n", len(data), name)
}

func main() {
	parseCmdArgs()

	tracer := trace.GetTracerProvider()

	_ = tracer

	kodosrv := kodo.NewStorageService(kodo.WithCredential(_accessKey, _secretKey), kodo.WithBucket(_bucket), kodo.WithEndpoint("http://ria8j59xt.hd-bkt.clouddn.com"))
	ctx := context.TODO()

	options := []oss.ListOption{kodo.WithListLimit(_limit)}
	if _prefix != "" {
		options = append(options, kodo.WithListPrefix(_prefix))
	}
	files, err := kodosrv.List(ctx, options...)
	if err != nil {
		util.Fatal(err)
	}
	for _, file := range files {
		fmt.Printf("file=%+v\n", file)

		func() {
			start := time.Now()
			defer func() { fmt.Printf("time elapsed for downing file %s: %v\n", file.Name, time.Since(start)) }()

		}()
	}
}
