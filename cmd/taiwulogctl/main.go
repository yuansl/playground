// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-11-13 10:35:05

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"bufio"
	"bytes"
	"compress/flate"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"path"
	"regexp"
	"runtime"
	"time"

	"github.com/qbox/net-deftones/clients/sinkv2"
	netutil "github.com/qbox/net-deftones/util"
	"golang.org/x/sync/errgroup"

	"github.com/yuansl/playground/logger"
	"github.com/yuansl/playground/oss"
	"github.com/yuansl/playground/oss/kodo"
	"github.com/yuansl/playground/util"
)

func downloadlogs(ctx context.Context, domain string, begin, end time.Time, outputdir string, taiwu TaiwuService) {
	egroup, ctx := errgroup.WithContext(ctx)

	for datetime := begin; datetime.Before(end); datetime = datetime.Add(5 * time.Minute) {
		links, err := taiwu.LogLink(ctx, domain, datetime)
		if err != nil {
			switch {
			case errors.Is(err, ErrInvalid):
				logger.FromContext(ctx).Warnf("taiwu.LogLink error: %v\n", err)
				return
			default:
				util.Fatal("loglink(domain=%s,datetime=%v): %v\n", domain, datetime, err)
			}
		}

		_datetime := datetime
		log := logger.FromContext(ctx)

		for i, link := range links {
			_i := i
			_link := link

			egroup.Go(func() error {
				log.Infof("downloading %q ...\n", _link)

				filename := path.Join(outputdir, fmt.Sprintf("/%s_%s-%04d.json", domain, _datetime.Format("200601021504"), _i))
				fp, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY|os.O_TRUNC, 0644)
				if err != nil {
					util.Fatal(err)
				}
				defer fp.Close()

				if err := netutil.WithRetry(ctx, func() error {
					if err := download(ctx, _link.Url, fp); err != nil {
						switch {
						case errors.Is(err, ErrInvalid):
							panic("BUG: you shoud review your request becuase of the error: " + err.Error())
						default:
							var cause flate.CorruptInputError
							if errors.As(err, &cause) {
								logger.FromContext(ctx).Warnf("download %q error: %v, skip ...\n", _link.Url, err)
								return nil
							}
							return err
						}
					}
					return nil
				}); err != nil {
					util.Fatal("download error:", err)
				}

				log.Infof("saved %q as %q \n", _link.Url, filename)
				return nil
			})
		}
	}
	egroup.Wait()
}

type TaiwuLog struct {
	Did    string
	Events []struct {
		Vvid       string
		Url        string
		Type       string
		Timeseries []struct {
			Timestamp int64 `json:"ts"`
			Cdn       int64
			P2p       int64
			Period    int64
		} `json:"flow"`
	} `json:"download"`
	processTime time.Time `json:"-"`
}

type TrafficPoint struct {
	Timestamp time.Time
	Bytes     int64
}

const (
	KiB = 1024
	MiB = 1024 * KiB
	GiB = 1024 * MiB
	TiB = 1024 * GiB
	PiB = 1024 * TiB
)

func prettyOfBytes(bytes float64) string {
	if gib := bytes / GiB; gib >= 1 {
		return fmt.Sprintf("%.2f GiB", gib)
	}
	if mib := bytes / MiB; mib >= 1 {
		return fmt.Sprintf("%.2f MiB", mib)
	}
	if kib := bytes / KiB; kib >= 1 {
		return fmt.Sprintf("%.2f KiB", kib)
	}
	return fmt.Sprintf("%.0f B", bytes)
}

var _filenameTimestampRegex = regexp.MustCompile(`[[:digit:]]{12}`)

type TaiwuStandardLog struct {
	Url       string
	Type      string
	Timestamp int64 `json:"ts"`
	Period    int64
	Cdn       int64
	P2p       int64
}

type pattern struct {
	domain    string
	timestamp time.Time
}

func saveAs(ctx context.Context, logs []TaiwuStandardLog, timestamp time.Time, uniq map[pattern]int, oss oss.ObjectStorageService) error {
	outfile, err := os.CreateTemp("/tmp/", "pdntaiwu_"+_domain+"*")
	if err != nil {
		return fmt.Errorf("os.CreateTemp: %v", err)
	}
	defer outfile.Close()

	buffered := bufio.NewWriter(outfile)

	gz := gzip.NewWriter(buffered)
	for _, it := range logs {
		if err = json.NewEncoder(gz).Encode(it); err != nil {
			return fmt.Errorf("json.Encode: %w", err)
		}
	}
	gz.Close()
	if err = buffered.Flush(); err != nil {
		return fmt.Errorf("bufio.Flush: %w", err)
	}
	timestamp = time.Date(timestamp.Year(), timestamp.Month(), timestamp.Day(), timestamp.Hour(), 0, 0, 0, timestamp.Location())
	key := pattern{domain: _domain, timestamp: timestamp}
	id, exists := uniq[key]
	if exists {
		id++
	}
	uniq[key] = id

	fkey := fmt.Sprintf("pdntaiwu/%s/%s/part-%04d.gz", timestamp.Format("2006-01-02-15"), _domain, id)

	outfile.Seek(0, io.SeekStart)

	logger.FromContext(ctx).Infof("uploading as key %s ...\n", fkey)

	_, err = oss.Upload(ctx, "defy-etl-log", outfile, kodo.WithKey(fkey))
	return err
}

type groupKey struct {
	Domain    string
	Day       time.Time
	Timestamp time.Time
}

type logline struct {
	bytes      []byte
	processDay time.Time
	file       string
}

type traffic struct {
	Domain    string
	Timestamp time.Time
	Bytes     int64
}

func aggregate(ctx context.Context, filenames []string, sinker TrafficSinker) error {
	var perTimestampP2P = make(map[groupKey]int64)
	var lineq = make(chan *logline, _concurrency)
	var taiwulogq = make(chan *TaiwuLog, _concurrency)
	var done = make(chan bool)

	go func() {
		start := time.Now()
		for taiwulog := range taiwulogq {
			for _, event := range taiwulog.Events {
				u, err := url.Parse(event.Url)
				if err != nil {
					logger.FromContext(ctx).Warnf("WARN: url.Parse(%q) error: %v, skipped\n", event.Url, err)
					continue
				}
				domain := u.Host

				for _, it := range event.Timeseries {
					eventTime := time.Unix(it.Timestamp/300_000*300_000/1000, 0)
					if eventTime.Before(taiwulog.processTime) || eventTime.Sub(taiwulog.processTime) >= 24*time.Hour {
						continue
					}
					day := time.Date(eventTime.Year(), eventTime.Month(), eventTime.Day(), 0, 0, 0, 0, eventTime.Location())

					perTimestampP2P[groupKey{Domain: domain, Timestamp: eventTime, Day: day}] += it.P2p
				}
			}
		}
		fmt.Printf("time elapsed for aggregate: %v\n", time.Since(start))
		done <- true
	}()

	go func() {
		defer close(taiwulogq)
		egroup1, ctx := errgroup.WithContext(ctx)
		egroup1.SetLimit(_concurrency)

		start := time.Now()

		for i := 0; i < _concurrency; i++ {
			egroup1.Go(func() error {
				for line := range lineq {
					var log TaiwuLog

					if err := json.Unmarshal(line.bytes, &log); err != nil {
						logger.FromContext(ctx).Infof("WARN: json.Unmarshal(content=`%s`): %v, file=%s\n", line, err, line.file)
						continue
					}
					log.processTime = line.processDay
					taiwulogq <- &log
				}
				return nil
			})
		}
		egroup1.Wait()

		fmt.Printf("time elapsed for json unmarshal: %v\n", time.Since(start))
	}()

	egroup0, ctx := errgroup.WithContext(ctx)

	start := time.Now()

	for _, file := range filenames {
		_file := file
		egroup0.Go(func() error {
			fp, err := os.Open(_file)
			if err != nil {
				util.Fatal(err)
			}
			defer fp.Close()

			logger.FromContext(ctx).Infof("aggregating file %s ...\n", _file)

			filecreateTime, err := time.ParseInLocation("200601021504", _filenameTimestampRegex.FindString(_file), time.Local)
			if err != nil {
				panic("BUG: file name pattern changed")
			}
			processTime := time.Date(filecreateTime.Year(), filecreateTime.Month(), filecreateTime.Day(), 0, 0, 0, 0, filecreateTime.Location())
			for r := bufio.NewReader(fp); ; {
				line, err := r.ReadBytes('\n')
				if err != nil {
					if !errors.Is(err, io.EOF) {
						util.Fatal("bufio.ReadBytes: %v\n", err)
					}
					break
				}
				lineq <- &logline{bytes: bytes.TrimSpace(line), file: _file, processDay: processTime}
			}
			return nil
		})
	}
	egroup0.Wait()
	fmt.Printf("time elapsed for reading from file: %v\n", time.Since(start))
	close(lineq)

	<-done

	var perTimestampTraffic = make(map[groupKey][]TrafficPoint)

	for group, bytes := range perTimestampP2P {
		key := groupKey{Domain: group.Domain, Day: group.Day}
		perTimestampTraffic[key] = append(perTimestampTraffic[key], TrafficPoint{Timestamp: group.Timestamp, Bytes: bytes})
	}

	var traffics []TrafficStat

	for key, timeseries := range perTimestampTraffic {
		traffics = append(traffics, TrafficStat{Domain: key.Domain, Region: RegionChina, Day: key.Day, Timeseries: timeseries})
	}
	if _sink {
		return sinker.Sink(context.WithoutCancel(ctx), traffics)
	}
	return nil
}

const (
	_ACCESS_KEY_DEFAULT  = "557TpseUM8ovpfUhaw8gfa2DQ0104ZScM-BTIcBx"
	_SECRET_KEY_DEFAULT  = "d9xLPyreEG59pR01sRQcFywhm4huL-XEpHHcVa90"
	_KODO_BUCKET_DEFAULT = "fusionlogtest"
)

var (
	_accessKey     = _ACCESS_KEY_DEFAULT
	_secretKey     = _SECRET_KEY_DEFAULT
	_bucket        = _KODO_BUCKET_DEFAULT
	_prefix        string
	_limit         int
	_begin         time.Time
	_end           time.Time
	_domain        string
	_outputdir     string
	_mode          string
	_robotsinkAddr string
	_version       int
	_concurrency   int
	_sink          bool
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
	flag.TextVar(&_begin, "begin", time.Time{}, "begin time")
	flag.TextVar(&_end, "end", time.Time{}, "end time")
	flag.StringVar(&_domain, "domain", "audiosdk.xmcdn.com", "specify cdn domain")
	flag.StringVar(&_outputdir, "dir", "/tmp", "directory for saving logs")
	flag.StringVar(&_mode, "mode", "stat", "specify mode(one of stat|download)")
	flag.StringVar(&_bucket, "bucket", _KODO_BUCKET_DEFAULT, "specify bucket name")
	flag.StringVar(&_prefix, "prefix", "", "specify a bucket key prefix")
	flag.StringVar(&_robotsinkAddr, "robotsink", "http://xs321:30060", "address of robotsink service")
	flag.IntVar(&_version, "version", 1, "api version")
	flag.IntVar(&_concurrency, "c", runtime.NumCPU(), "concurrency")
	flag.BoolVar(&_sink, "sink", false, "sink result if set")
	flag.Parse()
}

func main() {
	parseCmdArgs()

	go func() { http.ListenAndServe(":6060", nil) }()

	ctx := logger.NewContext(context.TODO(), logger.New())
	switch _mode {
	case "stat":
		client, err := sinkv2.NewClient(_robotsinkAddr)
		if err != nil {
			util.Fatal("sinkv2.NewClient:", err)
		}
		trafficSrv := NewTrafficSinker(client)

		if err := aggregate(ctx, flag.Args(), trafficSrv); err != nil {
			util.Fatal("aggregate:", err)
		}
	case "download":
		taiwucli := NewTaiwuClient()
		downloadlogs(ctx, _domain, _begin, _end, _outputdir, taiwucli)
	default:
		fmt.Fprintf(os.Stderr, `Usage: %s -mode <download|stat> -begin <RFC3339 time> `, os.Args[0])
		os.Exit(0)
	}
}
