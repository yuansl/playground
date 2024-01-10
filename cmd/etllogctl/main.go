package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	fusionutil "github.com/qbox/net-deftones/fusionrobot/util"
	"github.com/yuansl/playground/oss/kodo"
	"github.com/yuansl/playground/util"
	"golang.org/x/sync/errgroup"
)

type DnlivestreamLog struct {
	Outgoingbandwidth []int64 `json:"outgoingbandwidth"`
	RequestTime       int64   `json:"starttime"`
	Timestamp         int64   `json:"-"`
	BytesSent         int64   `json:"-"`
}

var _options struct {
	accessKey, secretKey string
	linkdomain           string
	prefix               string
	bucket               string
}

func init() {
	_options.accessKey = os.Getenv("ACCESS_KEY")
	_options.secretKey = os.Getenv("SECRET_KEY")
}

func parseCmdOptions() {
	flag.StringVar(&_options.linkdomain, "linkdomain", "http://ria8j59xt.hd-bkt.clouddn.com", "specify linkdomain of kodo")
	flag.StringVar(&_options.prefix, "prefix", "dnlivestream/2023-11-12-19/qn-pcdngw.cdn.huya.com/", "specify kodo object file prefix")
	flag.StringVar(&_options.bucket, "bucket", "defy-etl-log", "specify bucket name")
	flag.Parse()
}

func aligned5minTime(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, t.Hour(), (t.Minute()/5)*5, 0, 0, t.Location())
}

type lockedMap struct {
	m    map[time.Time]int64
	lock sync.Mutex
}

func statEtlLogFile(r io.Reader, groupby *lockedMap) error {
	for r := bufio.NewReader(r); ; {
		line, err := r.ReadBytes('\n')
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return fmt.Errorf("bufio.NewReader.Read: %w", err)
			}
			break
		}
		line = bytes.TrimSpace(line)

		var stdlog DnlivestreamLog

		if err = json.Unmarshal(line, &stdlog); err != nil {
			fmt.Printf("json.Unmarshal error: %v, skipped ...\n", err)
			continue
		}
		ts := aligned5minTime(time.Unix(stdlog.RequestTime/1000, 0))
		bytes := fusionutil.Sum(stdlog.Outgoingbandwidth)
		groupby.lock.Lock()
		groupby.m[ts] += bytes
		groupby.lock.Unlock()
	}
	return nil
}

func main() {
	parseCmdOptions()

	storage := kodo.NewStorageService(kodo.WithCredential(_options.accessKey, _options.secretKey), kodo.WithLinkDomain(_options.linkdomain))
	files, err := storage.List(context.TODO(), _options.bucket, kodo.WithListPrefix(_options.prefix))
	if err != nil {
		util.Fatal(err)
	}

	var groupby = lockedMap{m: make(map[time.Time]int64)} // time.Time -> int64
	var wg errgroup.Group
	for _, f := range files {
		_f := f
		wg.Go(func() error {
			fmt.Printf("stat file '%s' ...\n", _f.Name)
			data, err := storage.Download(context.TODO(), _options.bucket, _f.Name)
			if err != nil {
				util.Fatal(err)
			}
			gz, err := gzip.NewReader(bytes.NewReader(data))
			if err != nil {
				util.Fatal(err)
			}
			defer gz.Close()
			return statEtlLogFile(gz, &groupby)
		})
	}
	wg.Wait()

	fmt.Printf("traffic in bytes stat:\n")
	for ts, v := range groupby.m {
		fmt.Printf("timestamp: %s, bytes: %v\n", ts, v)
	}
}
