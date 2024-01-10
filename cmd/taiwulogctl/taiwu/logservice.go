package taiwu

import (
	"context"
	"time"
)

type Link struct {
	Url string
}

type LogService interface {
	LogLinks(ctx context.Context, domain string, timestamp time.Time, token ...string) ([]Link, error)
}
