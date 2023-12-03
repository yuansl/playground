package taiwu

import (
	"context"
	"time"
)

type Link struct {
	Url string
}

type LogService interface {
	LogLink(ctx context.Context, domain string, timestamp time.Time) ([]Link, error)
}
