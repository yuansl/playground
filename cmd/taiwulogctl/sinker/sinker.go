package sinker

import (
	"context"
	"time"
)

type TrafficSinker interface {
	Sink(ctx context.Context, _ []TrafficStat) error
}

type TrafficPoint struct {
	Timestamp time.Time
	Bytes     int64
}

type TrafficStat struct {
	Timeseries []TrafficPoint
	Day        time.Time
	Domain     string
	Region     Region
	DataType   DataType
}
