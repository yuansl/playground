package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	"github.com/qbox/net-deftones/logger"
	"golang.org/x/sync/errgroup"

	"github.com/yuansl/playground/cmd/taiwulogctl/sinker"
	"github.com/yuansl/playground/util"
)

const BUFSIZE = 1 << 20

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type TaiwuRawLog struct {
	File   string
	Did    string
	Domain string
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
}

type TaiwuStandardLog struct {
	File      string
	Domain    string
	Url       string
	Type      string
	Timestamp int64 `json:"ts"`
	Period    int64
	Cdn       int64
	P2p       int64
}

type logline struct {
	bytes []byte
	file  string
}

type ProcessWindow struct {
	Begin time.Time
	End   time.Time
}

type GroupKey struct {
	Domain    string
	Timestamp time.Time
}

type (
	TrafficSinker = sinker.TrafficSinker
	TrafficStat   = sinker.TrafficStat
	TrafficPoint  = sinker.TrafficPoint
)

const (
	RegionChina = sinker.RegionChina
)

var _logFilenameRegexp = regexp.MustCompile(`[[:alnum:]\.-_]+_`)

func extractDomainFrom(filename string) string {
	match := _logFilenameRegexp.Find(unsafe.Slice(unsafe.StringData(filename), len(filename)))
	_off := strings.LastIndex(unsafe.String(unsafe.SliceData(match), len(match)), "_")
	match = match[:_off]
	return filepath.Base(unsafe.String(unsafe.SliceData(match), len(match)))
}

func dayOf(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func stat(ctx context.Context, filenames []string, win ProcessWindow, sinker TrafficSinker) error {
	var lineq = make(chan *logline, _options.concurrency)
	var taiwuRawLogchan = make(chan *TaiwuRawLog, _options.concurrency)
	var taiwuStdLogschan = make(chan []TaiwuStandardLog, _options.concurrency)
	var linesCounter atomic.Int32

	start := time.Now()
	log := logger.FromContext(ctx)

	go func() {
		for range time.Tick(2 * time.Second) {
			log.Infof("read %d lines, since %v, elapsed time: %v\n", linesCounter.Load(), start, time.Since(start))
		}
	}()

	go func() {
		egroup, ctx := errgroup.WithContext(ctx)

		for _, file := range filenames {
			_file := file
			egroup.Go(func() error {
				fp, err := os.Open(_file)
				if err != nil {
					util.Fatal(err)
				}
				defer fp.Close()

				logger.FromContext(ctx).Infof("aggregating file %s ...\n", _file)

				for r := bufio.NewReaderSize(fp, BUFSIZE); ; {
					line, err := r.ReadBytes('\n')
					if err != nil {
						if !errors.Is(err, io.EOF) {
							util.Fatal("bufio.ReadBytes: %v\n", err)
						}
						break
					}
					lineq <- &logline{bytes: bytes.TrimSpace(line), file: _file}
				}
				return nil
			})
		}
		egroup.Wait()
		close(lineq)
	}()

	go func() {
		egroup, ctx := errgroup.WithContext(ctx)

		for i := 0; i < _options.concurrency; i++ {
			egroup.Go(func() error {
				for line := range lineq {
					linesCounter.Add(+1)

					var taiwulog TaiwuRawLog

					if err := json.Unmarshal(line.bytes, &taiwulog); err != nil {
						logger.FromContext(ctx).Infof("WARN: json.Unmarshal(content=`%s`) error: %v (file=%s, skip ...)\n",
							line.bytes, err, line.file)
						return nil
					}
					taiwulog.Domain = extractDomainFrom(line.file)
					taiwulog.File = line.file

					taiwuRawLogchan <- &taiwulog
				}
				return nil
			})
		}
		egroup.Wait()
		close(taiwuRawLogchan)
	}()

	go func() {
		egroup, _ := errgroup.WithContext(ctx)

		for i := 0; i < _options.concurrency; i++ {
			egroup.Go(func() error {
				for rawlog := range taiwuRawLogchan {
					for _, event := range rawlog.Events {
						var logs = make([]TaiwuStandardLog, 0, len(event.Timeseries))

						for _, it := range event.Timeseries {
							logs = append(logs, TaiwuStandardLog{
								File:      rawlog.File,
								Domain:    rawlog.Domain,
								Url:       event.Url,
								Type:      event.Type,
								Timestamp: it.Timestamp,
								Period:    it.Period,
								P2p:       it.P2p,
								Cdn:       it.Cdn,
							})
						}
						taiwuStdLogschan <- logs
					}
				}
				return nil
			})
		}
		egroup.Wait()
		close(taiwuStdLogschan)
	}()

	var (
		traffics        []TrafficStat
		perDayTraffic   = make(map[GroupKey][]TrafficPoint)
		perTimestampP2P = make(map[GroupKey]int64)
	)

	for stdlogs := range taiwuStdLogschan {
		for _, stdlog := range stdlogs {
			eventTime := time.Unix(stdlog.Timestamp/300_000*300_000/1000, 0)
			if eventTime.Before(win.Begin) || eventTime.Compare(win.End) >= 0 {
				continue
			}
			groupby := GroupKey{
				Domain:    stdlog.Domain,
				Timestamp: eventTime,
			}
			perTimestampP2P[groupby] += stdlog.P2p

		}
	}

	for groupBy, bytes := range perTimestampP2P {
		key := GroupKey{
			Domain:    groupBy.Domain,
			Timestamp: dayOf(groupBy.Timestamp),
		}
		perDayTraffic[key] = append(perDayTraffic[key], TrafficPoint{Timestamp: groupBy.Timestamp, Bytes: bytes})

		fmt.Printf("timestamp: %+v, bytes: %d\n", groupBy, bytes)
	}

	for groupBy, timeseries := range perDayTraffic {
		traffics = append(traffics, TrafficStat{Domain: groupBy.Domain, Region: RegionChina, Day: groupBy.Timestamp, Timeseries: timeseries})
	}

	return sinker.Sink(context.WithoutCancel(ctx), traffics)
}
