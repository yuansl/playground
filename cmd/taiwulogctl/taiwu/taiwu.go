package taiwu

import (
	"context"
	"time"
)

type Link struct {
	Url string
}

type TaiwuService interface {
	LogLink(ctx context.Context, domain string, timestamp time.Time) ([]Link, error)
}
