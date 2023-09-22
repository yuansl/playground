package bucket

import (
	"context"
	"time"
)

type BucketService interface {
	GetBucket(ctx context.Context, domain string, time time.Time) (string, error)
}
