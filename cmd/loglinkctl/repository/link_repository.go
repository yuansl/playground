package repository

import (
	"context"
	"time"
)

type LogLink struct {
	Id        string `bson:"_id"`
	Domain    string
	Hour      string
	Name      string
	Size      int64
	Url       string
	Mtime     int64
	Traffic   int64     `bson:"flux"`
	Timestamp time.Time `bson:"-"`
}

//go:generate stringer -type Business -linecomment
type Business int

const (
	BusinessCdn Business = iota // cdn
	BusinessSrc                 // src
)

type LinkOptions struct {
	Domain   string
	Url      string
	Business Business
	Filter   func(*LogLink) bool
}

type LinkRepository interface {
	SetDownloadUrl(ctx context.Context, link *LogLink, url string) error
	GetLinks(ctx context.Context, begin time.Time, end time.Time, opts ...*LinkOptions) ([]LogLink, error)
	DeleteLinks(ctx context.Context, links ...LogLink) error
}
