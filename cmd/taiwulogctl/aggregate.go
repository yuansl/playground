package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"net/url"
	"os"
	"sync/atomic"
	"time"

	jsoniter "github.com/json-iterator/go"
	"golang.org/x/sync/errgroup"

	"github.com/yuansl/playground/cmd/taiwulogctl/sinker"
	"github.com/yuansl/playground/logger"
	"github.com/yuansl/playground/util"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type TaiwuRawLog struct {
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
}

type TaiwuStandardLog struct {
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

func aggregate(ctx context.Context, filenames []string, w ProcessWindow, sinker TrafficSinker) error {
	var perTimestampP2P = make(map[GroupKey]int64)
	var lineq = make(chan *logline, _concurrency)
	var taiwuRawLogchan = make(chan *TaiwuRawLog, _concurrency)
	var taiwuStdLogchan = make(chan *TaiwuStandardLog, _concurrency)
	var done = make(chan bool)
	var linesCounter atomic.Int32

	start := time.Now()
	log := logger.FromContext(ctx)

	go func() {
		for range time.Tick(2 * time.Second) {
			log.Infof("read %d lines, since %v, elapsed time: %v\n", linesCounter.Load(), start, time.Since(start))
		}
	}()

	go func() {
		for stdlog := range taiwuStdLogchan {
			eventTime := time.Unix(stdlog.Timestamp/300_000*300_000/1000, 0)
			if eventTime.Before(w.Begin) || eventTime.Compare(w.End) >= 0 {
				continue
			}
			groupby := GroupKey{
				Domain:    stdlog.Domain,
				Timestamp: eventTime,
			}
			perTimestampP2P[groupby] += stdlog.P2p
		}
		done <- true
	}()

	go func() {
		defer close(taiwuStdLogchan)

		egroup, _ := errgroup.WithContext(ctx)
		for i := 0; i < _concurrency; i++ {
			egroup.Go(func() error {
				for rawlog := range taiwuRawLogchan {
					for i := 0; i < len(rawlog.Events); i++ {
						event := &rawlog.Events[i]
						u, err := url.Parse(event.Url)
						if err != nil {
							log.Warnf("WARN: url.Parse(%q) error: %v, skipped\n", event.Url, err)
							continue
						}
						for _, it := range event.Timeseries {
							taiwuStdLogchan <- &TaiwuStandardLog{
								Domain:    u.Host,
								Url:       event.Url,
								Type:      event.Type,
								Timestamp: it.Timestamp,
								Period:    it.Period,
								P2p:       it.P2p,
								Cdn:       it.Cdn,
							}
						}
					}
				}
				return nil
			})
		}
		egroup.Wait()
	}()

	go func() {
		defer close(taiwuRawLogchan)

		egroup, ctx := errgroup.WithContext(ctx)

		for i := 0; i < _concurrency; i++ {
			egroup.Go(func() error {
				for line := range lineq {
					linesCounter.Add(+1)

					var taiwulog TaiwuRawLog
					if err := json.Unmarshal(line.bytes, &taiwulog); err != nil {
						logger.FromContext(ctx).Infof("WARN: json.Unmarshal(content=`%s`) error: %v (file=%s, skip ...)\n", line.bytes, err, line.file)
						return nil
					}

					taiwuRawLogchan <- &taiwulog
				}
				return nil
			})
		}
		egroup.Wait()
	}()

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

			for r := bufio.NewReader(fp); ; {
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

	<-done

	var (
		traffics      []TrafficStat
		perDayTraffic = make(map[GroupKey][]TrafficPoint)
	)
	for groupBy, bytes := range perTimestampP2P {
		key := GroupKey{
			Domain:    groupBy.Domain,
			Timestamp: time.Date(groupBy.Timestamp.Year(), groupBy.Timestamp.Month(), groupBy.Timestamp.Day(), 0, 0, 0, 0, groupBy.Timestamp.Location()),
		}
		perDayTraffic[key] = append(perDayTraffic[key], TrafficPoint{Timestamp: groupBy.Timestamp, Bytes: bytes})
	}
	for groupBy, timeseries := range perDayTraffic {
		traffics = append(traffics, TrafficStat{Domain: groupBy.Domain, Region: RegionChina, Day: groupBy.Timestamp, Timeseries: timeseries})
	}

	return sinker.Sink(context.WithoutCancel(ctx), traffics)
}
