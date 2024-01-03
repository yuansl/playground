package repository

import (
	"context"
	"time"
)

type LogLink struct {
	Id     string `bson:"_id"`
	Domain string
	Hour   string
	Name   string
	Size   int64
	Url    string
	Mtime  int64
}

//go:generate stringer -type Business -linecomment
type Business int

const (
	BusinessCdn Business = iota // cdn
	BusinessSrc                 // src
)

type LinkRepository interface {
	SetDownloadUrl(ctx context.Context, link *LogLink, url string) error
	GetLinks(ctx context.Context, domain string, begin, end time.Time, _ Business) ([]LogLink, error)
}
